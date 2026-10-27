#!/usr/bin/env node
"use strict";

// Downloads the release binary for this platform, checks it against the
// sha256 in checksums.txt, and installs it into ./bin.
//
// checksums.txt is copied into the npm package by release.yml at publish
// time, so the expected hashes come from the (immutable, provenance-signed)
// npm tarball, not from the GitHub release the archive is downloaded from.
//
// Runs as the npm postinstall script, and from run.js on first use, because
// npm 12 skips install scripts by default. All output goes to stderr: when
// run.js calls this, stdout is the MCP stdio channel.

const { execFileSync } = require("child_process");
const crypto = require("crypto");
const fs = require("fs");
const os = require("os");
const path = require("path");
const https = require("https");

const VERSION = require("./package.json").version;
const REPO = "GeiserX/lynxprompt-mcp";
const BIN_NAME = process.platform === "win32" ? "lynxprompt-mcp.exe" : "lynxprompt-mcp";
const BIN_DIR = path.join(__dirname, "bin");
const BIN_PATH = path.join(BIN_DIR, BIN_NAME);
const CHECKSUMS_PATH = path.join(__dirname, "checksums.txt");

const MAX_REDIRECTS = 5;
const REQUEST_TIMEOUT_MS = 30_000;
const MAX_DOWNLOAD_BYTES = 64 * 1024 * 1024;
// GitHub release downloads redirect from github.com to its asset host.
const ALLOWED_HOSTS = ["github.com", "objects.githubusercontent.com", "release-assets.githubusercontent.com"];

function log(msg) {
  process.stderr.write(`lynxprompt-mcp: ${msg}\n`);
}

function getPlatformArch(platform = process.platform, arch = process.arch) {
  const platformMap = { darwin: "darwin", linux: "linux", win32: "windows" };
  const archMap = { x64: "amd64", arm64: "arm64" };
  const p = platformMap[platform];
  const a = archMap[arch];
  if (!p || !a) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }
  return { platform: p, arch: a };
}

function getAssetName(version = VERSION, platform, arch) {
  const pa = getPlatformArch(platform, arch);
  const ext = pa.platform === "windows" ? "zip" : "tar.gz";
  return `lynxprompt-mcp_${version}_${pa.platform}_${pa.arch}.${ext}`;
}

// expectedHash returns the sha256 for assetName from checksums.txt content.
function expectedHash(checksumsText, assetName) {
  for (const line of checksumsText.split(/\r?\n/)) {
    const m = line.trim().match(/^([0-9a-f]{64})\s+\*?(.+)$/i);
    if (m && m[2] === assetName) {
      return m[1].toLowerCase();
    }
  }
  throw new Error(`no sha256 for ${assetName} in checksums.txt`);
}

// checkURL allows only https URLs on the release hosts.
function checkURL(raw) {
  let u;
  try {
    u = new URL(raw);
  } catch {
    throw new Error(`refusing malformed download URL`);
  }
  if (u.protocol !== "https:") {
    throw new Error(`refusing non-https download URL (${u.protocol})`);
  }
  if (!ALLOWED_HOSTS.includes(u.hostname)) {
    throw new Error(`refusing download from unexpected host ${u.hostname}`);
  }
  return u;
}

// download streams url to dest, following https redirects on the allowed
// hosts, and resolves with the sha256 of what it wrote.
function download(url, dest, redirects = 0) {
  return new Promise((resolve, reject) => {
    let u;
    try {
      u = checkURL(url);
    } catch (err) {
      return reject(err);
    }
    const req = https.get(u, { timeout: REQUEST_TIMEOUT_MS }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        res.resume();
        if (redirects >= MAX_REDIRECTS) {
          return reject(new Error(`too many redirects (>${MAX_REDIRECTS})`));
        }
        const next = new URL(res.headers.location, u).toString();
        return download(next, dest, redirects + 1).then(resolve, reject);
      }
      if (res.statusCode !== 200) {
        res.resume();
        return reject(new Error(`download failed: HTTP ${res.statusCode}`));
      }
      const hash = crypto.createHash("sha256");
      const out = fs.createWriteStream(dest, { mode: 0o600 });
      let size = 0;
      res.on("data", (chunk) => {
        size += chunk.length;
        if (size > MAX_DOWNLOAD_BYTES) {
          req.destroy(new Error(`download larger than ${MAX_DOWNLOAD_BYTES} bytes`));
          return;
        }
        hash.update(chunk);
      });
      res.on("error", reject);
      req.on("error", (err) => {
        out.destroy();
        reject(err);
      });
      out.on("error", reject);
      out.on("finish", () => resolve(hash.digest("hex")));
      res.pipe(out);
    });
    req.on("timeout", () => req.destroy(new Error(`request timed out after ${REQUEST_TIMEOUT_MS}ms`)));
    req.on("error", reject);
  });
}

// extract unpacks only the binary from the archive into dir, without a shell.
// tar reads both .tar.gz and, on Windows 10+, .zip.
function extract(archive, dir) {
  execFileSync("tar", ["-xf", archive, "-C", dir, BIN_NAME], { stdio: "ignore" });
}

async function install() {
  if (fs.existsSync(BIN_PATH)) {
    return BIN_PATH;
  }
  if (!fs.existsSync(CHECKSUMS_PATH)) {
    throw new Error("checksums.txt is missing from the package, so the download cannot be verified; reinstall lynxprompt-mcp from npm");
  }
  const assetName = getAssetName();
  const want = expectedHash(fs.readFileSync(CHECKSUMS_PATH, "utf8"), assetName);
  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${assetName}`;

  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), "lynxprompt-mcp-"));
  try {
    const archive = path.join(tmp, assetName);
    log(`downloading v${VERSION} for ${process.platform}-${process.arch}`);
    const got = await download(url, archive);
    if (got !== want) {
      throw new Error(`sha256 mismatch for ${assetName}: expected ${want}, got ${got}`);
    }
    extract(archive, tmp);
    const extracted = path.join(tmp, BIN_NAME);
    if (process.platform !== "win32") {
      fs.chmodSync(extracted, 0o755);
    }
    fs.mkdirSync(BIN_DIR, { recursive: true });
    // Move into place in one step so a half-written binary is never run.
    const staged = path.join(BIN_DIR, `.${BIN_NAME}.${process.pid}`);
    fs.copyFileSync(extracted, staged);
    fs.renameSync(staged, BIN_PATH);
    log(`installed ${BIN_PATH} (sha256 verified)`);
    return BIN_PATH;
  } finally {
    fs.rmSync(tmp, { recursive: true, force: true });
  }
}

module.exports = { install, expectedHash, checkURL, getAssetName, BIN_PATH };

if (require.main === module) {
  install().catch((err) => {
    log(`install failed: ${err.message}`);
    process.exit(1);
  });
}
