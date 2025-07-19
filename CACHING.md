# Report Caching

The my-day CLI now includes intelligent report caching to improve performance and reduce LLM API calls.

## Overview

Reports are automatically cached based on a unique fingerprint that includes:
- Report date
- Configuration parameters (LLM settings, debug flags, etc.)
- Issue data (keys and update times)
- Comment data (IDs and timestamps)
- Worklog data

This ensures that cached reports are only reused when the underlying data hasn't changed.

## Cache Commands

### Generate Reports with Caching

By default, caching is enabled for all report generation:

```bash
# Normal report generation (with caching)
my-day report

# Filter to show only tickets updated in the last 48 hours
my-day report --since 48h

# Filter to show only tickets updated in the last 3 days
my-day report --since 72h

# Disable caching for this report
my-day report --no-cache

# Use only cached reports (fail if no cache exists)
my-day report --cache-only
```

### Export Cached Reports

Export previously generated reports without calling the LLM:

```bash
# List all cached reports
my-day export --list

# Export today's report
my-day export --date 2025-01-15

# Export reports for a date range
my-day export --from 2025-01-10 --to 2025-01-15

# Export to specific directory with custom template
my-day export --output-dir ./reports --filename-template "standup_{{.Date}}"
```

### Cache Management

```bash
# List cached reports with details
my-day cache list

# Show cache statistics
my-day cache stats

# Clear all cached reports
my-day cache clear --all

# Clear reports older than a date
my-day cache clear --before 2025-01-01

# Delete specific cached reports
my-day cache delete 2025-01-15_abcd1234
```

## Cache Storage

Reports are cached in `~/.my-day/reports/` with the following structure:

- `index.json` - Cache index with metadata
- `<report-id>.json` - Individual cached reports

Each cached report includes:
- Generated report content
- Configuration used
- Generation metadata (time, LLM usage, etc.)
- Input data fingerprint

## Cache Benefits

1. **Performance**: Instant report generation for unchanged data
2. **Cost Savings**: Reduces LLM API calls
3. **Offline Access**: View reports without network connectivity
4. **Export Flexibility**: Export reports in different formats without regeneration

## Data Filtering with --since

The `--since` flag filters cached data to include only tickets updated within the specified time period:

- `--since 24h` - Last 24 hours
- `--since 48h` - Last 48 hours  
- `--since 72h` - Last 3 days
- `--since 168h` - Last 7 days (default)

This filtering happens on the locally cached data, so it's very fast and doesn't require new API calls.

## Cache Invalidation

Cache is automatically invalidated when:
- Issue data changes (updates, new comments)
- Configuration changes (LLM settings, format, etc.)
- Report parameters change (debug flags, grouping, `--since` value, etc.)

## Debug Information

Use debug flags to see cache activity:

```bash
# Show cache hits/misses
my-day report --debug

# Show detailed cache information
my-day report --verbose
```

## Examples

```bash
# Generate report for last 48 hours (caches for future use)
my-day report --since 48h --format markdown

# Export the same report without regeneration
my-day export --date today --format markdown

# Generate report for specific date with custom time range
my-day report --date 2025-01-15 --since 72h

# List all cached reports
my-day cache list

# Export last week's reports
my-day export --from 2025-01-08 --to 2025-01-14 --output-dir ./weekly-reports
```