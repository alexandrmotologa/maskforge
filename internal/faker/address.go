package faker

import (
	"fmt"
)

var streetNames = []string{
	"Main Street", "High Street", "Maple Avenue", "Oak Street", "Washington Boulevard",
	"Park Avenue", "Cedar Road", "Pine Lane", "Elm Street", "Sunset Drive",
	"Broadway", "Market Street", "Church Street", "Lincoln Way", "Victoria Street",
	"Strada Victoriei", "Bulevardul Unirii", "Strada Mihai Eminescu", "Calea Floreasca",
	"Friedrichstrasse", "Rue de la Paix", "Gran Via", "Corso Vittorio Emanuele",
}

var cities = []string{
	"New York", "San Francisco", "London", "Bucharest", "Berlin", "Paris", "Amsterdam",
	"Madrid", "Rome", "Toronto", "Sydney", "Cluj-Napoca", "Timisoara", "Iasi",
	"Munich", "Frankfurt", "Dublin", "Zurich", "Stockholm", "Copenhagen", "Vienna",
	"Warsaw", "Prague", "Tokyo", "Singapore", "Austin", "Seattle", "Chicago", "Boston",
}

var countries = []string{
	"United States", "Romania", "United Kingdom", "Germany", "France",
	"Spain", "Italy", "Netherlands", "Canada", "Australia",
	"Sweden", "Switzerland", "Poland", "Japan", "Austria",
}

var companies = []string{
	"Apex Systems", "BluePeak Technologies", "Crestline Logistics", "Datacraft Labs",
	"EchoWave Networks", "Frontier Data", "Greenfield Capital", "Harbor Solutions",
	"Ironclad Security", "Jupiter Dynamics", "Krypton Systems", "Luminary Software",
	"Matrix Industrial", "Nexus Analytics", "Orion Cloud", "Pinnacle Media",
	"Quantum Stream", "Radiant Global", "Starlight Digital", "Terraform Ventures",
}

// GenerateStreetAddress generates a deterministic street address.
func GenerateStreetAddress(rng *DeterministicPRNG) string {
	num := rng.IntBetween(1, 999)
	street := rng.PickString(streetNames)
	return fmt.Sprintf("%d %s", num, street)
}

// GenerateCity returns a deterministic city name.
func GenerateCity(rng *DeterministicPRNG) string {
	return rng.PickString(cities)
}

// GenerateCountry returns a deterministic country name.
func GenerateCountry(rng *DeterministicPRNG) string {
	return rng.PickString(countries)
}

// GenerateZipCode returns a deterministic 5 or 6 digit postal code.
func GenerateZipCode(rng *DeterministicPRNG) string {
	return fmt.Sprintf("%05d", rng.IntBetween(10000, 99999))
}

// GenerateCompany returns a deterministic company name.
func GenerateCompany(rng *DeterministicPRNG) string {
	return rng.PickString(companies)
}
