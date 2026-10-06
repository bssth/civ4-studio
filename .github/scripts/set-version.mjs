// Writes the release version into wails.json (info.productVersion), which Wails uses for
// the Windows file properties and the macOS Info.plist. Windows accepts only numbers,
// so "v1.2.3-beta.1" becomes "1.2.3"; anything that is not a version becomes "0.0.0".
// Usage: node .github/scripts/set-version.mjs v1.2.3
import {readFileSync, writeFileSync} from 'node:fs';

const tag = process.argv[2] ?? '';
const match = /^v?(\d+)\.(\d+)\.(\d+)/.exec(tag);
const version = match ? `${match[1]}.${match[2]}.${match[3]}` : '0.0.0';

const config = JSON.parse(readFileSync('wails.json', 'utf8'));
config.info = {...config.info, productVersion: version};
writeFileSync('wails.json', JSON.stringify(config, null, 2) + '\n');
console.log(`productVersion: ${version} (from "${tag}")`);
