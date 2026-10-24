// Generates PWA icon PNGs (public/pwa-192.png, public/pwa-512.png) and the
// maskable variants: a solid blue rounded square with a white plate circle and
// a yellow leaf stroke. Run once: node scripts/make-icons.mjs
import { deflateSync } from "node:zlib";
import { writeFileSync, mkdirSync } from "node:fs";

function crc32(buf) {
  let table = crc32.table;
  if (!table) {
    table = crc32.table = new Int32Array(256);
    for (let n = 0; n < 256; n++) {
      let c = n;
      for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
      table[n] = c;
    }
  }
  let c = 0xffffffff;
  for (const b of buf) c = table[(c ^ b) & 0xff] ^ (c >>> 8);
  return (c ^ 0xffffffff) >>> 0;
}

function chunk(type, data) {
  const len = Buffer.alloc(4);
  len.writeUInt32BE(data.length);
  const body = Buffer.concat([Buffer.from(type), data]);
  const crc = Buffer.alloc(4);
  crc.writeUInt32BE(crc32(body));
  return Buffer.concat([len, body, crc]);
}

function png(size, pixels /* RGBA flat buffer */) {
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(size, 0);
  ihdr.writeUInt32BE(size, 4);
  ihdr[8] = 8; // bit depth
  ihdr[9] = 6; // RGBA
  const raw = Buffer.alloc(size * (size * 4 + 1));
  for (let y = 0; y < size; y++) {
    raw[y * (size * 4 + 1)] = 0; // filter: none
    pixels.copy(raw, y * (size * 4 + 1) + 1, y * size * 4, (y + 1) * size * 4);
  }
  return Buffer.concat([
    Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    chunk("IHDR", ihdr),
    chunk("IDAT", deflateSync(raw)),
    chunk("IEND", Buffer.alloc(0)),
  ]);
}

// Simple shapes: rounded-square background, white plate ring, yellow leaf arc.
function draw(size) {
  const px = Buffer.alloc(size * size * 4);
  const r = size / 24; // draw in a 24x24 coordinate space
  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      const [dx, dy] = [x / r + 0.0001, y / r + 0.0001];
      let cr = 0x00, cg = 0x72, cb = 0xb2, a = 0xff; // #0072b2 background
      // rounded corners: cut the four 24r squares' corners (radius 4)
      const rr = 4;
      const cx = Math.min(dx, 24 - dx), cy = Math.min(dy, 24 - dy);
      if (cx < rr && cy < rr && Math.hypot(rr - cx, rr - cy) > rr) a = 0x00;
      // plate ring centered (10.5,13) r 4.6 stroke 1.6
      let dplate = Math.abs(Math.hypot(dx - 10.5, dy - 13) - 4.6);
      if (dplate <= 0.8) { cr = cg = cb = 0xff; }
      // inner dot r1.4
      if (Math.hypot(dx - 10.5, dy - 13) <= 1.4) { cr = cg = cb = 0xff; }
      // leaf stroke: arc from (13.5,18.2) curving up-right, approximated by the segment of circle center (11,12.5) radius 8.7 between angles
      const angle = Math.atan2(dy - 12.5, dx - 11);
      const dleaf = Math.abs(Math.hypot(dx - 11, dy - 12.5) - 8.7);
      if (dplate > 0.8 && dleaf <= 0.8 && angle > -0.9 && angle < 0.9 && dx > 12 && dy > 10 && dy < 19) {
        cr = 0xf0; cg = 0xe4; cb = 0x42;
      }
      // leaf fill blade
      if (dx >= 12 && dx <= 17.5 && dy >= 6 && dy <= 11.2 && dx <= 17.5 - (dy - 6) * 1.4) {
        cr = 0xf0; cg = 0xe4; cb = 0x42;
      }
      const i = (y * size + x) * 4;
      px[i] = cr; px[i + 1] = cg; px[i + 2] = cb; px[i + 3] = a;
    }
  }
  return px;
}

mkdirSync("public", { recursive: true });
for (const size of [192, 512]) {
  writeFileSync(`public/pwa-${size}.png`, png(size, draw(size)));
  writeFileSync(`public/pwa-${size}-maskable.png`, png(size, draw(size)));
}
console.log("icons written");
