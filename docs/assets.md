# Assets command

The `assets` command helps prepare image assets for web and application projects. It provides these subcommands:

- `assets favicon` — generates a favicon set from a single source image
- `assets resize` — resizes a source image to a target width and height
- `assets convert` — converts an image between PNG, JPG/JPEG, GIF, and WebP
- `assets config init` — creates the default asset configuration file
- `assets pdf` — creates a multipage PDF from numbered PNG, JPG/JPEG, and WebP images

The command group itself only displays help when run without a subcommand.

## Command tree

```bash
cmdr assets --help
cmdr assets favicon --help
cmdr assets resize --help
cmdr assets convert --help
cmdr assets config --help
cmdr assets config init --help
cmdr assets pdf --help
```

## `assets favicon`

Generates a favicon bundle from one source image and writes PNG files into the selected output directory.

### Usage

```bash
cmdr assets favicon --input path/to/source.png --output ./generated
cmdr assets favicon -i ./logo.png -o ./generated --sizes 16,32,48,180,192,512
cmdr assets favicon -i ./logo.png -o ./generated --overwrite
```

### Flags

- `--input`, `-i`: required source image path
- `--output`, `-o`: output directory for generated files; if omitted, the configured default is used
- `--sizes`: comma-separated values such as `16,32,48`; when omitted, the command uses the configured favicon preset or the default set `[16, 32, 48, 180, 192, 512]`
- `--overwrite`: allow overwriting existing files in the target output directory

### Behavior

- The source image is decoded and validated before processing.
- The command creates a staging directory inside the output folder and only finalizes files after generation succeeds.
- The default file names are:
  - `favicon-16x16.png`
  - `favicon-32x32.png`
  - `favicon-48x48.png`
  - `apple-touch-icon.png` for `180`
  - `icon-192.png` for `192`
  - `icon-512.png` for `512`
- Generated favicon sizes are always PNG files.
- Duplicate values are removed and sorted numerically.
- Each value must be a positive integer and cannot exceed `2048`.

### Example

```bash
cmdr assets favicon --input ./assets/logo.png --output ./public --sizes 16,32,48,192 --overwrite
```

This creates PNG favicon files in `./public`, including `favicon-16x16.png`, `favicon-32x32.png`, `favicon-48x48.png`, and `icon-192.png`.

## `assets resize`

Resizes a source image to a target width and height using one of the supported fit modes.

### Usage

```bash
cmdr assets resize --input ./source.png --output ./out/resized.png --width 640 --height 480
cmdr assets resize -i ./source.png -o ./out/resized.jpg -w 300 -H 200 --fit cover
cmdr assets resize -i ./source.png -o ./out/resized.gif -w 256 -H 256 --fit stretch --overwrite
```

### Flags

- `--input`, `-i`: required source image path
- `--output`, `-o`: required output image path
- `--width`, `-w`: required target width; must be greater than zero
- `--height`, `-H`: required target height; must be greater than zero
- `--fit`: resize mode; accepted values are `contain`, `cover`, and `stretch`; default is `contain`
- `--overwrite`: replace an existing output file when set

### Behavior

- The source and output path must differ; the command rejects identical paths.
- The output format is inferred from the destination file extension.
- Supported output formats are `png`, `jpg`, `jpeg`, `gif`, and `webp`.
- If the destination extension is `.jpg` or `.jpeg`, the command will handle alpha-containing input by compositing it over white before encoding.
- WebP output preserves alpha and uses the shared WebP encoder.
- If output file already exists and `--overwrite` is not provided, the command returns an error.

### Fit modes

- `contain`: preserve aspect ratio and fit the image inside the target rectangle while keeping empty padding if needed
- `cover`: preserve aspect ratio and cover the target rectangle without distortion
- `stretch`: resize directly to the requested dimensions, potentially distorting the image

## `assets convert`

Converts an image to a supported destination format. The destination format is inferred from the output extension.

### Usage

```bash
cmdr assets convert -i logo.png -o logo.webp
cmdr assets convert -i photo.jpg -o photo.webp
cmdr assets convert -i photo.webp -o photo.png
cmdr assets convert -i photo.webp -o photo.jpg
cmdr assets convert -i photo.png -o photo.webp --quality 85
cmdr assets convert -i photo.png -o photo.webp --overwrite
cmdr assets convert -i photo.png -o photo.webp --lossless
```

### Flags

- `--input`, `-i`: required source image path
- `--output`, `-o`: required destination image path
- `--format`: optional explicit format; it must match the output extension
- `--quality`, `-q`: WebP quality from 1 to 100; default 85
- `--lossless`: encode WebP without loss
- `--overwrite`: replace an existing output file

`--quality` and `--lossless` are ignored by neither PNG/JPEG/GIF encoders: the command rejects those options for non-WebP output. PNG and WebP preserve transparency. JPEG and GIF output follow existing image composition behavior; JPEG alpha is composited on white. GIF input is decoded as a single frame and is not converted as an animated GIF/WebP.

## `assets config`

Manages the asset configuration file used by the command group.

### `assets config init`

Creates the default asset configuration file in the standard location.

### Usage

```bash
cmdr assets config init
cmdr assets config init --force
cmdr assets config init --path ./custom/assets.json
```

### Flags

- `--force`, `-f`: overwrite an existing configuration file
- `--path`, `-p`: custom config path; default is `~/.cmdr/assets.json`

### Default configuration

The generated file contains:

```json
{
  "defaults": {
    "output_dir": "generated",
    "format": "png",
    "overwrite": false
  },
  "presets": {
    "favicon": {
      "sizes": [16, 32, 48, 180, 192, 512],
      "format": "png",
      "preserve_transparency": true
    }
  }
}
```

The config file is loaded with default values when it is missing. Supported default and preset formats are `png`, `jpg`, `jpeg`, `gif`, and `webp`.

## `assets pdf`

Creates one PDF page per numbered image found directly in a source directory (or recursively with `--recursive`). The generated PDF is written beside the images by default.

### Usage

```bash
cmdr assets pdf ./ZeroDayZ7
cmdr assets pdf ./ZeroDayZ7 --output numer-001.pdf
cmdr assets pdf ./ZeroDayZ7 --page-size A4
cmdr assets pdf ./ZeroDayZ7 --orientation portrait
cmdr assets pdf ./ZeroDayZ7 --margin 5
cmdr assets pdf ./ZeroDayZ7 --force
cmdr assets pdf ./ZeroDayZ7 --recursive
```

Supported page sizes are `A4` and `LETTER`. Supported orientations are `auto`, `portrait`, and `landscape`; `auto` uses each image's proportions. The default margin is `0` mm.

### Numeric filename rules

- Supported names start with digits, for example `1.png`, `001_okladka.png`, or `010_final.jpg`.
- Sorting uses the numeric prefix; extensions do not affect order.
- Values such as `1`, `2`, `3`, `10` and `001`, `002`, `010` are sorted numerically.
- Two files with the same numeric prefix are rejected because their order would be ambiguous.
- Files without a numeric prefix are skipped and reported.

### Flags

- `--output`, `-o`: custom output PDF name or path
- `--page-size`: `A4` or `LETTER` (default `A4`)
- `--orientation`: `auto`, `portrait`, or `landscape` (default `auto`)
- `--margin`: margin in millimeters (default `0`)
- `--force`: allow replacing an existing PDF
- `--recursive`: include supported images in subdirectories

By default, images maintain their aspect ratio and are contained within the usable page area without cropping or stretching. PNG transparency is decoded and composited safely. Source images are never modified or optimized.

## Supported formats

The asset system accepts these output formats:

- `png`
- `jpg`
- `jpeg`
- `gif`
- `webp`

For `favicon`, the generated files are always PNG images. For `resize`, the destination extension determines the output format. For `convert`, the output format is based on the specified output file extension.

## Notes

- If `--sizes` is omitted for `assets favicon`, the command tries to use the configured `favicon` preset and falls back to the default list if no preset is available.
- `assets` exposes `favicon`, `resize`, `convert`, and `config init` subcommands.
- When a command writes files, it protects against accidental overwrite by requiring `--overwrite` for existing files unless the operation is explicitly allowed by the command logic.
- GIF decoding uses the existing single-image decoder. Animated GIF frames are not preserved as animated WebP or other animated output.
