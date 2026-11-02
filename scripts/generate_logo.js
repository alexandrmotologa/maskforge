const fs = require('fs');
const path = require('path');
const { Resvg } = require('@resvg/resvg-js');

function buildLogoSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="1024" height="1024">
  <defs>
    <clipPath id="squircle-clip">
      <rect x="24" y="24" width="976" height="976" rx="220" />
    </clipPath>

    <linearGradient id="cyan-glow" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#00f5ff"/>
      <stop offset="100%" stop-color="#0284c7"/>
    </linearGradient>

    <linearGradient id="amber-forge" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#fbbf24"/>
      <stop offset="100%" stop-color="#ea580c"/>
    </linearGradient>

    <linearGradient id="titanium-plate" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#475569"/>
      <stop offset="50%" stop-color="#334155"/>
      <stop offset="100%" stop-color="#1e293b"/>
    </linearGradient>

    <linearGradient id="obsidian-shade" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#1e293b"/>
      <stop offset="100%" stop-color="#090d16"/>
    </linearGradient>

    <filter id="subtle-shadow" x="-10%" y="-10%" width="120%" height="120%">
      <feDropShadow dx="0" dy="16" stdDeviation="20" flood-color="#000000" flood-opacity="0.14" />
    </filter>
  </defs>

  <!-- Luxury White Squircle Container -->
  <rect x="24" y="24" width="976" height="976" rx="220" fill="#ffffff" stroke="#e2e8f0" stroke-width="6" />

  <g clip-path="url(#squircle-clip)">
    <g transform="translate(512, 512)" filter="url(#subtle-shadow)">

      <!-- Hexagonal Architectural Gateway Frame -->
      <polygon points="
        0,-390
        338,-195
        338,195
        0,390
        -338,195
        -338,-195
      " fill="none" stroke="#0f172a" stroke-width="36" stroke-linejoin="round" />

      <polygon points="
        0,-355
        307,-177
        307,177
        0,355
        -307,177
        -307,-177
      " fill="none" stroke="#00f5ff" stroke-width="4" opacity="0.45" stroke-dasharray="16, 12" />

      <!-- Corner Alignment Crosshairs -->
      <line x1="-338" y1="-195" x2="-308" y2="-177" stroke="#00f5ff" stroke-width="3" opacity="0.6" />
      <line x1="338" y1="-195" x2="308" y2="-177" stroke="#00f5ff" stroke-width="3" opacity="0.6" />
      <line x1="0" y1="-390" x2="0" y2="-355" stroke="#00f5ff" stroke-width="3" opacity="0.6" />
      <line x1="0" y1="390" x2="0" y2="355" stroke="#00f5ff" stroke-width="3" opacity="0.6" />

      <!-- ============================================================== -->
      <!-- 100% Watertight Solid Base Silhouette (Zero-Leak Rule)          -->
      <!-- ============================================================== -->
      <path d="
        M 0,-240
        L 75,-210
        L 155,-325
        L 168,-365
        L 150,-310
        L 215,-175
        L 235,-65
        L 265,35
        L 190,135
        L 220,285
        L 110,340
        L 0,365
        L -110,340
        L -220,285
        L -190,135
        L -265,35
        L -235,-65
        L -215,-175
        L -150,-310
        L -168,-365
        L -155,-325
        L -75,-210
        Z
      " fill="#090d16" />

      <!-- ============================================================== -->
      <!-- Ear Anatomy & Feathers (Lynx Tufts)                           -->
      <!-- ============================================================== -->
      <!-- Left Ear Outer & Inner -->
      <polygon points="-75,-210 -155,-325 -215,-175" fill="#1e293b" />
      <polygon points="-90,-215 -148,-310 -185,-180" fill="#0f172a" />
      <polygon points="-110,-220 -145,-295 -165,-190" fill="#00f5ff" opacity="0.18" />
      <line x1="-155" y1="-325" x2="-168" y2="-365" stroke="#00f5ff" stroke-width="5" stroke-linecap="round" />
      <line x1="-150" y1="-320" x2="-158" y2="-355" stroke="#38bdf8" stroke-width="3" stroke-linecap="round" />

      <!-- Right Ear Outer & Inner -->
      <polygon points="75,-210 155,-325 215,-175" fill="#334155" />
      <polygon points="90,-215 148,-310 185,-180" fill="#1e293b" />
      <polygon points="110,-220 145,-295 165,-190" fill="#00f5ff" opacity="0.25" />
      <line x1="155" y1="-325" x2="168" y2="-365" stroke="#00f5ff" stroke-width="5" stroke-linecap="round" />
      <line x1="150" y1="-320" x2="158" y2="-355" stroke="#38bdf8" stroke-width="3" stroke-linecap="round" />

      <!-- ============================================================== -->
      <!-- Cranium & Forehead Low-Poly Facets                             -->
      <!-- ============================================================== -->
      <polygon points="0,-240 -75,-210 -65,-140 0,-165" fill="#1e293b" />
      <polygon points="0,-240 75,-210 65,-140 0,-165" fill="#334155" />
      <polygon points="0,-165 -65,-140 0,-85" fill="#0f172a" />
      <polygon points="0,-165 65,-140 0,-85" fill="#1e293b" />

      <!-- Temple & Cheek Flairs (Lynx Ruffs) -->
      <polygon points="-75,-210 -215,-175 -235,-65 -145,-90" fill="#1e293b" />
      <polygon points="75,-210 215,-175 235,-65 145,-90" fill="#475569" />
      <polygon points="-235,-65 -265,35 -175,25 -145,-90" fill="#0f172a" />
      <polygon points="235,-65 265,35 175,25 145,-90" fill="#334155" />
      <polygon points="-265,35 -190,135 -125,75 -175,25" fill="#1e293b" />
      <polygon points="265,35 190,135 125,75 175,25" fill="#475569" />

      <!-- ============================================================== -->
      <!-- The Forged Mask / Armor Visor (MaskForge Identity Plate)       -->
      <!-- ============================================================== -->
      <!-- Central Forged Visor Bridge -->
      <polygon points="0,-85 -70,-60 -50,15 0,35" fill="#334155" />
      <polygon points="0,-85 70,-60 50,15 0,35" fill="#475569" />

      <!-- Visor Wing Left & Right -->
      <polygon points="-70,-60 -145,-90 -160,-25 -75,0 -50,15" fill="#1e293b" />
      <polygon points="70,-60 145,-90 160,-25 75,0 50,15" fill="#334155" />

      <!-- Visor Cyan Accent Trim -->
      <polyline points="-160,-25 -70,-60 0,-85 70,-60 160,-25" fill="none" stroke="#00f5ff" stroke-width="3.5" stroke-linecap="round" stroke-linejoin="round" />
      <polyline points="-145,-90 -70,-60 -50,15 0,35 50,15 70,-60 145,-90" fill="none" stroke="#00f5ff" stroke-width="2" opacity="0.6" stroke-linejoin="round" />

      <!-- ============================================================== -->
      <!-- Predator Optics (Almond Slanted Eyes, Upward Cant)             -->
      <!-- ============================================================== -->
      <!-- Left Eye Socket & Eye -->
      <polygon points="-125,-42 -65,-30 -85,-18 -135,-28" fill="#05080f" />
      <polygon points="-120,-38 -72,-28 -88,-20 -128,-26" fill="url(#cyan-glow)" />
      <!-- Left Pupil Slit -->
      <polygon points="-98,-34 -94,-34 -92,-22 -96,-22" fill="#ffffff" />

      <!-- Right Eye Socket & Eye -->
      <polygon points="125,-42 65,-30 85,-18 135,-28" fill="#05080f" />
      <polygon points="120,-38 72,-28 88,-20 128,-26" fill="url(#cyan-glow)" />
      <!-- Right Pupil Slit -->
      <polygon points="94,-34 98,-34 96,-22 92,-22" fill="#ffffff" />

      <!-- ============================================================== -->
      <!-- Muzzle, Whisker Pads & Jaw (Refined Slate, No Buck Teeth)      -->
      <!-- ============================================================== -->
      <!-- Nose Leather Keystone -->
      <polygon points="0,55 -22,40 22,40" fill="#090d16" stroke="#00f5ff" stroke-width="1.5" />

      <!-- Whisker Pads Spreading Laterally -->
      <polygon points="-22,40 0,55 -35,95 -80,65 -50,15" fill="#334155" />
      <polygon points="22,40 0,55 35,95 80,65 50,15" fill="#64748b" />
      <polygon points="0,55 0,105 -35,95" fill="#1e293b" />
      <polygon points="0,55 0,105 35,95" fill="#475569" />

      <!-- Chin & Lower Jaw -->
      <polygon points="0,105 -35,95 -45,150 0,175" fill="#0f172a" />
      <polygon points="0,105 35,95 45,150 0,175" fill="#1e293b" />

      <!-- ============================================================== -->
      <!-- Chest Mantle & Forged Core Medallion                           -->
      <!-- ============================================================== -->
      <!-- Collar Left & Right Plates -->
      <polygon points="0,175 -45,150 -125,75 -190,135 -140,230 0,260" fill="#1e293b" />
      <polygon points="0,175 45,150 125,75 190,135 140,230 0,260" fill="#334155" />

      <!-- Shoulders & Lower Armor -->
      <polygon points="-140,230 -190,135 -220,285 -110,340 0,260" fill="#0f172a" />
      <polygon points="140,230 190,135 220,285 110,340 0,260" fill="#1e293b" />

      <!-- Central Hexagonal Forge Core (Warm Amber Accent) -->
      <polygon points="
        0,275
        35,295
        35,335
        0,355
        -35,335
        -35,295
      " fill="#090d16" stroke="url(#amber-forge)" stroke-width="4" stroke-linejoin="round" />

      <polygon points="
        0,285
        22,298
        22,328
        0,342
        -22,328
        -22,298
      " fill="url(#amber-forge)" opacity="0.9" />

      <!-- Stream Piping Vector Lines emanating from Forge Core -->
      <line x1="-35" y1="315" x2="-80" y2="315" stroke="#00f5ff" stroke-width="3" opacity="0.6" stroke-dasharray="6,4" />
      <line x1="35" y1="315" x2="80" y2="315" stroke="#00f5ff" stroke-width="3" opacity="0.6" stroke-dasharray="6,4" />
      <line x1="0" y1="355" x2="0" y2="365" stroke="#00f5ff" stroke-width="3" opacity="0.6" />

    </g>
  </g>
</svg>`;
}

async function renderLogo() {
  const outputDir = path.join(__dirname, '..', 'docs', 'images');
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }

  const svg = buildLogoSvg();
  const svgPath = path.join(outputDir, 'logo.svg');
  const pngPath = path.join(outputDir, 'logo.png');

  fs.writeFileSync(svgPath, svg);
  const resvg = new Resvg(svg, { fitTo: { mode: 'width', value: 1024 } });
  fs.writeFileSync(pngPath, resvg.render().asPng());
  console.log('✓ Successfully rendered logo.svg and logo.png at 1024x1024 into docs/images');
}

renderLogo().catch((err) => {
  console.error('Error rendering logo:', err);
  process.exit(1);
});
