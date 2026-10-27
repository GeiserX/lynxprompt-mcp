"use strict";

// node --test postinstall.test.js (no network: every case fails before a download)
const test = require("node:test");
const assert = require("node:assert");
const fs = require("fs");
const os = require("os");
const path = require("path");

const { expectedHash, checkURL, getAssetName } = require("./postinstall.js");

const HASH = "a".repeat(64);

test("expectedHash finds the asset line", () => {
  const text = `${"b".repeat(64)}  lynxprompt-mcp_1.0.0_linux_arm64.tar.gz\n${HASH}  lynxprompt-mcp_1.0.0_linux_amd64.tar.gz\n`;
  assert.strictEqual(expectedHash(text, "lynxprompt-mcp_1.0.0_linux_amd64.tar.gz"), HASH);
});

test("expectedHash does not match a suffix of another asset name", () => {
  const text = `${HASH}  evil-lynxprompt-mcp_1.0.0_linux_amd64.tar.gz\n`;
  assert.throws(() => expectedHash(text, "lynxprompt-mcp_1.0.0_linux_amd64.tar.gz"), /no sha256/);
});

test("checkURL refuses http and foreign hosts", () => {
  assert.throws(() => checkURL("http://github.com/GeiserX/x"), /non-https/);
  assert.throws(() => checkURL("https://evil.example/x"), /unexpected host/);
  assert.throws(() => checkURL("not a url"), /malformed/);
  assert.strictEqual(checkURL("https://github.com/GeiserX/x").hostname, "github.com");
  assert.strictEqual(checkURL("https://release-assets.githubusercontent.com/x").protocol, "https:");
});

test("getAssetName matches the goreleaser archive names", () => {
  assert.strictEqual(getAssetName("0.2.0", "linux", "x64"), "lynxprompt-mcp_0.2.0_linux_amd64.tar.gz");
  assert.strictEqual(getAssetName("0.2.0", "darwin", "arm64"), "lynxprompt-mcp_0.2.0_darwin_arm64.tar.gz");
  assert.strictEqual(getAssetName("0.2.0", "win32", "x64"), "lynxprompt-mcp_0.2.0_windows_amd64.zip");
  assert.throws(() => getAssetName("0.2.0", "sunos", "x64"), /Unsupported/);
});

function isolatedCopy(withChecksums) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "lpmcp-test-"));
  fs.copyFileSync(path.join(__dirname, "postinstall.js"), path.join(dir, "postinstall.js"));
  fs.writeFileSync(path.join(dir, "package.json"), JSON.stringify({ version: "9.9.9" }));
  if (withChecksums !== undefined) {
    fs.writeFileSync(path.join(dir, "checksums.txt"), withChecksums);
  }
  return dir;
}

test("install refuses to download without the bundled checksums.txt", async () => {
  const dir = isolatedCopy(undefined);
  try {
    const { install } = require(path.join(dir, "postinstall.js"));
    await assert.rejects(install(), /checksums.txt is missing/);
  } finally {
    fs.rmSync(dir, { recursive: true, force: true });
  }
});

test("install refuses an asset that checksums.txt does not list", async () => {
  const dir = isolatedCopy(`${HASH}  something-else.tar.gz\n`);
  try {
    const { install } = require(path.join(dir, "postinstall.js"));
    await assert.rejects(install(), /no sha256/);
    assert.ok(!fs.existsSync(path.join(dir, "bin")), "nothing may be installed");
  } finally {
    fs.rmSync(dir, { recursive: true, force: true });
  }
});
