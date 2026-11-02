package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/alexandrmotologa/maskforge/internal/config"
	"github.com/alexandrmotologa/maskforge/internal/discovery"
	"github.com/alexandrmotologa/maskforge/internal/server"
	"github.com/alexandrmotologa/maskforge/internal/stream"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		// If piped data is available on STDIN, default to "run"
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			runPipe(os.Args[1:])
			return
		}
		printUsage()
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "run":
		runPipe(args)
	case "inspect":
		runInspect(args)
	case "studio", "serve":
		runStudio(args)
	case "benchmark", "bench":
		runBenchmark(args)
	case "version", "-v", "--version":
		fmt.Printf("MaskForge v%s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		fmt.Println("High-speed streaming database anonymizer and sensitive PII masker.")
	case "help", "-h", "--help":
		printUsage()
	default:
		// If first arg starts with a flag (e.g. -c or -i), assume "run"
		if strings.HasPrefix(command, "-") {
			runPipe(os.Args[1:])
			return
		}
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`MaskForge: Streaming Production Database Anonymizer & PII Masker

Usage:
  maskforge [command] [options]

Commands:
  run         Stream and mask SQL dump from STDIN to STDOUT or between files
  inspect     Scan database dump, audit PII risks, and auto-generate mask.yaml
  studio      Launch interactive web studio and visual rule builder (port 9880)
  benchmark   Run high-throughput synthetic stream benchmark
  version     Display version and build information

Examples:
  # Zero-disk pipe between pg_dump and psql:
  pg_dump prod_db | maskforge run --config mask.yaml | psql staging_db

  # Direct file-to-file transformation:
  maskforge run -c mask.yaml -i prod_dump.sql -o staging_masked.sql

  # Inspect database dump and generate configuration:
  maskforge inspect prod_dump.sql --generate-config mask.yaml

  # Launch visual web studio:
  maskforge studio --port 9880`)
}

func runPipe(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to mask.yaml configuration file")
	fs.StringVar(configPath, "c", "", "Path to mask.yaml configuration file (shorthand)")
	inputFile := fs.String("input", "", "Path to input SQL file (default: STDIN)")
	fs.StringVar(inputFile, "i", "", "Path to input SQL file (shorthand)")
	outputFile := fs.String("output", "", "Path to output SQL file (default: STDOUT)")
	fs.StringVar(outputFile, "o", "", "Path to output SQL file (shorthand)")
	quiet := fs.Bool("quiet", false, "Suppress progress and throughput statistics")
	fs.BoolVar(quiet, "q", false, "Suppress progress (shorthand)")

	_ = fs.Parse(args)

	var cfg *config.Config
	var err error

	if *configPath != "" {
		cfg, err = config.LoadFromFile(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration from %s: %v\n", *configPath, err)
			os.Exit(1)
		}
	} else {
		// Look for default mask.yaml in current directory
		if _, statErr := os.Stat("mask.yaml"); statErr == nil {
			cfg, err = config.LoadFromFile("mask.yaml")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading default mask.yaml: %v\n", err)
				os.Exit(1)
			}
			if !*quiet {
				fmt.Fprintln(os.Stderr, "Loaded configuration from mask.yaml")
			}
		} else {
			cfg = config.DefaultConfig()
			if !*quiet {
				fmt.Fprintln(os.Stderr, "No mask.yaml found; using default passthrough configuration.")
			}
		}
	}

	// Setup input reader
	var reader io.Reader = os.Stdin
	if *inputFile != "" {
		f, err := os.Open(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file %s: %v\n", *inputFile, err)
			os.Exit(1)
		}
		defer f.Close()
		reader = f
	}

	// Setup output writer
	var writer io.Writer = os.Stdout
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file %s: %v\n", *outputFile, err)
			os.Exit(1)
		}
		defer f.Close()
		writer = f
	}

	if !*quiet {
		fmt.Fprintf(os.Stderr, "MaskForge starting streaming engine... (salt: %s...)\n", cfg.Salt[:8])
	}

	pipeline := stream.NewPipeline(cfg)
	metrics, err := pipeline.ProcessStream(reader, writer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Streaming failed: %v\n", err)
		os.Exit(1)
	}

	if !*quiet {
		fmt.Fprintf(os.Stderr, "\nStreaming Complete:\n")
		fmt.Fprintf(os.Stderr, "  Lines Processed: %d\n", metrics.LinesProcessed)
		fmt.Fprintf(os.Stderr, "  Rows Masked:     %d\n", metrics.RowsMasked)
		fmt.Fprintf(os.Stderr, "  Data Streamed:   %.2f MB\n", float64(metrics.BytesRead)/(1024*1024))
		fmt.Fprintf(os.Stderr, "  Duration:        %s\n", metrics.Duration.Round(time.Millisecond))
		fmt.Fprintf(os.Stderr, "  Throughput:      %.0f rows/sec (%.2f MB/sec)\n", metrics.RowsPerSec, metrics.MBPerSec)
	}
}

func runInspect(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	genConfig := fs.String("generate-config", "", "Auto-generate mask.yaml to specified path")
	fs.StringVar(genConfig, "g", "", "Auto-generate config (shorthand)")
	samplesPerCol := fs.Int("samples", 10, "Maximum sample values to inspect per column")

	_ = fs.Parse(args)
	remaining := fs.Args()

	var reader io.Reader = os.Stdin
	sourceName := "STDIN"

	if len(remaining) > 0 && remaining[0] != "-" {
		filePath := remaining[0]
		f, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", filePath, err)
			os.Exit(1)
		}
		defer f.Close()
		reader = f
		sourceName = filePath
	}

	fmt.Fprintf(os.Stderr, "Auditing SQL dump %s for sensitive PII...\n\n", sourceName)

	scanner := discovery.NewSchemaScanner("")
	report, err := scanner.ScanReader(reader, *samplesPerCol)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Inspection failed: %v\n", err)
		os.Exit(1)
	}

	// Print summary audit table
	fmt.Printf("========================================================================================\n")
	fmt.Printf("                             MASKFORGE PII AUDIT REPORT                                \n")
	fmt.Printf("========================================================================================\n")
	fmt.Printf("Total Tables:     %d\n", report.TotalTables)
	fmt.Printf("Total Columns:    %d\n", report.TotalColumns)
	fmt.Printf("PII Columns:      %d\n", report.PIIColumns)
	fmt.Printf("Estimated Risk:   %s (Critical: %d, High: %d, Medium: %d, Low: %d)\n\n",
		report.EstimatedRisk, report.CriticalCount, report.HighCount, report.MediumCount, report.LowCount)

	for _, tbl := range report.Tables {
		fmt.Printf("TABLE: %s (Rows: %d)\n", tbl.Name, tbl.TotalRows)
		fmt.Printf("----------------------------------------------------------------------------------------\n")
		fmt.Printf("%-24s %-20s %-10s %-22s %s\n", "COLUMN", "TYPE", "RISK", "RECOMMENDED RULE", "COMPLIANCE")
		fmt.Printf("----------------------------------------------------------------------------------------\n")
		for _, col := range tbl.Columns {
			riskStr := string(col.Risk)
			compStr := strings.Join(col.Compliance, ", ")
			if compStr == "" {
				compStr = "-"
			}
			fmt.Printf("%-24s %-20s %-10s %-22s %s\n",
				truncate(col.Name, 24),
				truncate(col.DetectedType, 20),
				riskStr,
				truncate(col.RecommendedRule, 22),
				truncate(compStr, 24),
			)
		}
		fmt.Println()
	}

	if *genConfig != "" {
		cfg := discovery.GenerateConfigFromReport(report)
		yamlBytes, err := cfg.ToYAML()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error serializing configuration: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(*genConfig, yamlBytes, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config to %s: %v\n", *genConfig, err)
			os.Exit(1)
		}
		fmt.Printf("Successfully generated configuration at %s\n", *genConfig)
	}
}

func runStudio(args []string) {
	fs := flag.NewFlagSet("studio", flag.ExitOnError)
	port := fs.Int("port", 9880, "Port to run web studio server")
	fs.IntVar(port, "p", 9880, "Port shorthand")
	noBrowser := fs.Bool("no-browser", false, "Do not auto-open browser")

	_ = fs.Parse(args)

	srv := server.NewServer()
	srv.AttachStaticWeb()

	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost:%d", *port)

	fmt.Printf("===================================================================\n")
	fmt.Printf("                  MaskForge Interactive Web Studio                 \n")
	fmt.Printf("===================================================================\n")
	fmt.Printf("Studio running at:  %s\n", url)
	fmt.Printf("API Health:         %s/api/health\n", url)
	fmt.Printf("Press Ctrl+C to terminate.\n\n")

	if !*noBrowser {
		go func() {
			time.Sleep(300 * time.Millisecond)
			openBrowser(url)
		}()
	}

	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		fmt.Fprintf(os.Stderr, "Studio server failed: %v\n", err)
		os.Exit(1)
	}
}

func runBenchmark(args []string) {
	fs := flag.NewFlagSet("benchmark", flag.ExitOnError)
	numRows := fs.Int("rows", 100000, "Number of synthetic rows to benchmark")
	fs.IntVar(numRows, "n", 100000, "Number of rows shorthand")

	_ = fs.Parse(args)

	fmt.Printf("Benchmarking MaskForge streaming engine with %d synthetic rows...\n", *numRows)

	// Build synthetic PostgreSQL COPY stream
	var inBuf bytes.Buffer
	inBuf.WriteString("COPY public.users (id, email, first_name, last_name, phone, ssn, card_number) FROM stdin;\n")
	for i := 1; i <= *numRows; i++ {
		fmt.Fprintf(&inBuf, "%d\tuser%d@prod-domain.com\tAlice%d\tSmith%d\t+155501%04d\t123-45-6789\t4111111111111111\n",
			i, i, i, i, i%10000)
	}
	inBuf.WriteString("\\.\n")

	cfgYAML := `
version: "1.0"
salt: "benchmark-performance-salt"
tables:
  users:
    columns:
      email: "faker.email"
      first_name: "faker.first_name"
      last_name: "faker.last_name"
      phone: "faker.phone_number"
      ssn: "faker.ssn"
      card_number: "faker.credit_card"
`
	cfg, err := config.LoadFromBytes([]byte(cfgYAML))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Benchmark config load error: %v\n", err)
		os.Exit(1)
	}

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	var devNull bytes.Buffer
	devNull.Grow(inBuf.Len())

	start := time.Now()
	pipeline := stream.NewPipeline(cfg)
	metrics, err := pipeline.ProcessStream(&inBuf, &devNull)
	elapsed := time.Since(start)

	runtime.ReadMemStats(&m2)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Benchmark run error: %v\n", err)
		os.Exit(1)
	}

	allocMB := float64(m2.TotalAlloc-m1.TotalAlloc) / (1024 * 1024)

	fmt.Printf("\nBenchmark Results:\n")
	fmt.Printf("  Total Rows Masked: %d\n", metrics.RowsMasked)
	fmt.Printf("  Data Streamed:     %.2f MB\n", float64(metrics.BytesRead)/(1024*1024))
	fmt.Printf("  Duration:          %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  Throughput:        %.0f rows/sec\n", float64(metrics.RowsMasked)/elapsed.Seconds())
	fmt.Printf("  Stream Bandwidth:  %.2f MB/sec\n", (float64(metrics.BytesRead)/(1024*1024))/elapsed.Seconds())
	fmt.Printf("  RAM Allocated:     %.2f MB\n", allocMB)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
