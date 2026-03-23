#!/usr/bin/env node

import fs from 'node:fs/promises'
import path from 'node:path'

const defaultInput = 'D:/mudr/staging/mudrotop-source/raw/merged.json'
const defaultOutput = 'out/movie-catalog.slim.json'

function parseArgs(argv) {
  const args = new Map()

  for (let index = 0; index < argv.length; index += 1) {
    const current = argv[index]
    const next = argv[index + 1]

    if (!current.startsWith('--')) {
      continue
    }

    args.set(current.slice(2), next)
    index += 1
  }

  return {
    input: args.get('input') ?? process.env.MOVIE_CATALOG_SOURCE ?? defaultInput,
    output: args.get('output') ?? process.env.MOVIE_CATALOG_OUTPUT ?? defaultOutput,
  }
}

function normalizeGenre(raw) {
  return String(raw ?? '').trim().toLowerCase()
}

function toNumber(value) {
  const number = Number(value)
  return Number.isFinite(number) ? number : undefined
}

function buildId(item, index) {
  if (item?.id != null) {
    return String(item.id)
  }

  const seed = [item?.name, item?.alternativeName, item?.year, index + 1]
    .filter(Boolean)
    .join('-')
    .toLowerCase()
    .replace(/[^a-z0-9а-яё-]+/gi, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')

  return seed || `movie-${index + 1}`
}

function slimMovie(item, index) {
  const genres = Array.from(
    new Set(
      (Array.isArray(item?.genres) ? item.genres : [])
        .map((genre) => normalizeGenre(genre?.name ?? genre))
        .filter(Boolean),
    ),
  ).sort()

  if (!item?.name || genres.length === 0) {
    return null
  }

  return {
    id: buildId(item, index),
    name: String(item.name),
    alternative_name: item.alternativeName ? String(item.alternativeName) : undefined,
    year: toNumber(item.year),
    duration: toNumber(item.movieLength),
    rating: toNumber(item?.rating?.kp),
    poster_url: item?.poster?.url ?? item?.poster?.previewUrl ?? undefined,
    description: item?.description ?? item?.shortDescription ?? undefined,
    kp_url: item?.webUrl ?? item?.url ?? undefined,
    genres,
  }
}

async function main() {
  const { input, output } = parseArgs(process.argv.slice(2))

  const raw = await fs.readFile(input, 'utf8')
  const parsed = JSON.parse(raw)
  const movies = (Array.isArray(parsed) ? parsed : []).map(slimMovie).filter(Boolean)

  const payload = {
    generated_at: new Date().toISOString(),
    source: path.resolve(input),
    movies,
  }

  await fs.mkdir(path.dirname(output), { recursive: true })
  await fs.writeFile(output, JSON.stringify(payload, null, 2))
  console.log(`prepared ${movies.length} movies -> ${output}`)
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
})
