#!/usr/bin/env node

const { execFileSync } = require("child_process");
const path = require("path");
const os = require("os");
const fs = require("fs");

function getBinaryPath() {
  const platform = os.platform();
  const arch = os.arch();

  const platformMap = {
    darwin: "darwin",
    linux: "linux",
    win32: "win",
  };

  const archMap = {
    arm64: "arm64",
    x64: "x64",
    ia32: "x64", // fallback
  };

  const p = platformMap[platform];
  const a = archMap[arch];

  if (!p || !a) {
    console.error(`Unsupported platform: ${platform}-${arch}`);
    process.exit(1);
  }

  const pkgName = `@knowns/${p}-${a}`;
  const ext = platform === "win32" ? ".exe" : "";

  // Try to find in node_modules (installed as optionalDependency)
  const candidates = [
    // npm/pnpm standard layout
    path.join(__dirname, "..", "node_modules", pkgName, `knowns${ext}`),
    // pnpm hoisted
    path.join(__dirname, "..", "..", pkgName, `knowns${ext}`),
    // npm hoisted
    path.join(
      __dirname,
      "..",
      "..",
      "..",
      "node_modules",
      pkgName,
      `knowns${ext}`
    ),
    // Global install
    path.join(__dirname, "..", "..", pkgName, `knowns${ext}`),
  ];

  for (const candidate of candidates) {
    if (fs.existsSync(candidate)) {
      return candidate;
    }
  }

  // Try require.resolve as last resort
  try {
    const pkgDir = path.dirname(require.resolve(`${pkgName}/package.json`));
    const binary = path.join(pkgDir, `knowns${ext}`);
    if (fs.existsSync(binary)) {
      return binary;
    }
  } catch {}

  console.error(
    `Could not find knowns binary for ${platform}-${arch}.\n` +
      `Expected package: ${pkgName}\n` +
      `Try reinstalling: npm install knowns`
  );
  process.exit(1);
}

const binary = getBinaryPath();
const args = process.argv.slice(2);

try {
  execFileSync(binary, args, {
    stdio: "inherit",
    env: process.env,
  });
} catch (err) {
  if (err.status !== undefined) {
    process.exit(err.status);
  }
  throw err;
}
