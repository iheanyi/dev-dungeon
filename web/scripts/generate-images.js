#!/usr/bin/env node
/**
 * Generate PNG versions of SVG assets for social media and legacy browser support.
 * Run with: node scripts/generate-images.js
 */
import sharp from 'sharp';
import { readFileSync, writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const staticDir = join(__dirname, '..', 'static');

async function generateImages() {
	console.log('Generating PNG images from SVGs...\n');

	// OG Image (1200x630 for social media)
	console.log('Converting og-image.svg -> og-image.png');
	const ogSvg = readFileSync(join(staticDir, 'og-image.svg'));
	await sharp(ogSvg)
		.resize(1200, 630)
		.png()
		.toFile(join(staticDir, 'og-image.png'));
	console.log('  ✓ og-image.png (1200x630)');

	// Apple Touch Icon (180x180 is recommended)
	console.log('Converting favicon.svg -> apple-touch-icon.png');
	const faviconSvg = readFileSync(join(staticDir, 'favicon.svg'));
	await sharp(faviconSvg)
		.resize(180, 180)
		.png()
		.toFile(join(staticDir, 'apple-touch-icon.png'));
	console.log('  ✓ apple-touch-icon.png (180x180)');

	// Favicon ICO (32x32 PNG, browsers accept PNG for favicon)
	console.log('Converting favicon.svg -> favicon-32.png');
	await sharp(faviconSvg)
		.resize(32, 32)
		.png()
		.toFile(join(staticDir, 'favicon-32.png'));
	console.log('  ✓ favicon-32.png (32x32)');

	// Icon for PWA (192x192)
	console.log('Converting favicon.svg -> icon-192.png');
	await sharp(faviconSvg)
		.resize(192, 192)
		.png()
		.toFile(join(staticDir, 'icon-192.png'));
	console.log('  ✓ icon-192.png (192x192)');

	// Icon for PWA (512x512)
	console.log('Converting favicon.svg -> icon-512.png');
	await sharp(faviconSvg)
		.resize(512, 512)
		.png()
		.toFile(join(staticDir, 'icon-512.png'));
	console.log('  ✓ icon-512.png (512x512)');

	console.log('\n✓ All images generated successfully!');
}

generateImages().catch(console.error);
