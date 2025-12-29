# roadcraft-mod

A Go application that modifies vehicles in the game RoadCraft from Sabre Interactive.

## Features

- Reads and modifies PAK files (ZIP files with Store compression)
- Modifies `allowedPercent` values in `.cls` files
- Supports targeting specific files or modifying all `.cls` files
- Lists all `.cls` files in a PAK archive
- Preserves original file structure and content
- Supports both the actual CLS format and XML format (backwards compatibility)

## Installation

### From Source

```bash
go build -o roadcraft-mod .
```

## Usage

### Basic Usage

Modify all `.cls` files in a PAK to set `AllowedPercentage` to 0:

```bash
roadcraft-mod -pak /path/to/default_other.pak -percentage 0
```

### List Files

List all `.cls` files in a PAK:

```bash
roadcraft-mod -pak /path/to/default_other.pak -list
```

### Target Specific File

Modify only a specific file:

```bash
roadcraft-mod -pak /path/to/default_other.pak \
  -file ssl/autogen_designer_wizard/trucks/auto_ziks605e_mobile_scalper_res/auto_zikz_605e_mobile_scalper_res.cls \
  -percentage 0
```

### Save to Different File

Save modified PAK to a different location:

```bash
roadcraft-mod -pak /path/to/default_other.pak \
  -output /path/to/default_other_modified.pak \
  -percentage 0
```

## Command-Line Options

- `-pak <path>`: Path to the `default_other.pak` file (required)
- `-percentage <value>`: New allowedpercentage value (default: 0.0)
- `-file <path>`: Target specific `.cls` file within PAK (optional)
- `-output <path>`: Output path for modified PAK (default: overwrites input)
- `-list`: List all `.cls` files in the PAK

## Finding the PAK File

The `default_other.pak` file is typically located at:

**Windows:**
```
C:\Program Files (x86)\Steam\steamapps\common\RoadCraft\paks\client\default\default_other.pak
```

**Linux (Steam):**
```
~/.steam/steam/steamapps/common/RoadCraft/paks/client/default/default_other.pak
```

Or wherever your Steam library is installed under:
```
<SteamLibrary>/steamapps/common/RoadCraft/paks/client/default/default_other.pak
```

## Example: Allow Scraping Everywhere

To modify the ZIKZ 605E Mobile Scalper to allow scraping everywhere:

```bash
roadcraft-mod -pak "C:\Program Files (x86)\Steam\steamapps\common\RoadCraft\paks\client\default\default_other.pak" \
  -file ssl/autogen_designer_wizard/trucks/auto_ziks605e_mobile_scalper_res/auto_zikz_605e_mobile_scalper_res.cls \
  -percentage 0
```

**Note:** Always backup the original PAK file before modifying it.

## How It Works

1. Opens the PAK file (which is a ZIP archive with Store compression)
2. Finds `.cls` files matching the criteria
3. Modifies the `allowedPercent` value in the CLS property format
4. Preserves all other content in the files (spacing, structure, other properties)
5. Writes the modified PAK file back with Store compression

## CLS File Format

The `.cls` files use a custom property format:

```
properties = {
   prop_truck_mobile_sand_screen = {
      allowedPercent = 0.4
      belt1SpeedCoef = 0.3
      ...
   }
}
```

The tool modifies the `allowedPercent` field while preserving all other content and formatting.

## License

See [LICENSE](LICENSE) file for details.

