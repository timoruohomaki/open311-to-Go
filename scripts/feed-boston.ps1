<#
.SYNOPSIS
  Feed Boston 311 CSV rows (data.boston.gov export format) into the open311-to-Go
  API via the idempotent PUT /requests/{id} upsert endpoint.

.DESCRIPTION
  Maps each CSV row to Open311 JSON: standard fields top-level, jurisdiction
  extras under `properties` (see dictionaries/boston-311.yaml). The spatial-data-
  lake priority is coordinate preservation, so latitude/longitude are sent
  whenever they are valid (numeric, non-zero, inside a Boston bounding box); the
  server derives an indexed GeoJSON location from them. Rows with neither valid
  coordinates nor an address are skipped (the API requires a location).

  Re-runnable: PUT is keyed on service_request_id (Boston case_enquiry_id), so
  re-feeding updates in place rather than duplicating. Boston-local timestamps
  are converted to UTC (DST-aware). Respects the API's rate limit (default
  10/min) by pacing requests and honoring Retry-After on 429.

.PARAMETER CsvPath
  Path to the Boston 311 CSV export.

.PARAMETER ApiKey
  X-API-Key value. Defaults to $env:OPEN311_API_KEY. Never hard-code secrets.

.PARAMETER BaseUrl
  API base, e.g. https://api.ruohomaki.fi/open311/v2 (default).

.PARAMETER First / .PARAMETER Skip
  Take First rows after skipping Skip rows (for batching).

.PARAMETER DelaySeconds
  Pause between requests. Default 6.5s (~9.2/min, just under a 10/min limit).

.PARAMETER DryRun
  Map and print the JSON for the selected rows without sending anything.

.EXAMPLE
  $env:OPEN311_API_KEY = '...' ; ./feed-boston.ps1 -CsvPath C:\data\boston.csv -First 25
#>
[CmdletBinding()]
param(
  [Parameter(Mandatory = $true)] [string] $CsvPath,
  [string] $ApiKey = $env:OPEN311_API_KEY,
  [string] $BaseUrl = 'https://api.ruohomaki.fi/open311/v2',
  [int] $First = 25,
  [int] $Skip = 0,
  [double] $DelaySeconds = 6.5,
  [switch] $Bulk,
  [int] $BatchSize = 500,
  [switch] $DryRun
)

$ErrorActionPreference = 'Stop'
if (-not $DryRun -and [string]::IsNullOrWhiteSpace($ApiKey)) {
  throw "ApiKey is required (pass -ApiKey or set `$env:OPEN311_API_KEY). Use -DryRun to preview without a key."
}
if (-not (Test-Path $CsvPath)) { throw "CSV not found: $CsvPath" }

$eastern = [System.TimeZoneInfo]::FindSystemTimeZoneById('Eastern Standard Time')
$inv = [System.Globalization.CultureInfo]::InvariantCulture

function ConvertTo-UtcIso([string]$s) {
  if ([string]::IsNullOrWhiteSpace($s)) { return $null }
  $dt = [datetime]::MinValue
  if ([datetime]::TryParse($s, $inv, [System.Globalization.DateTimeStyles]::None, [ref]$dt)) {
    $local = [datetime]::SpecifyKind($dt, [System.DateTimeKind]::Unspecified)
    $utc = [System.TimeZoneInfo]::ConvertTimeToUtc($local, $eastern)
    return $utc.ToString("yyyy-MM-ddTHH:mm:ssZ")
  }
  return $null
}

function Get-ValidCoord([string]$latS, [string]$lonS) {
  $lat = 0.0; $lon = 0.0
  $okLat = [double]::TryParse($latS, [System.Globalization.NumberStyles]::Float, $inv, [ref]$lat)
  $okLon = [double]::TryParse($lonS, [System.Globalization.NumberStyles]::Float, $inv, [ref]$lon)
  if ($okLat -and $okLon -and $lat -ne 0 -and $lon -ne 0 -and
      $lat -gt 41 -and $lat -lt 43.5 -and $lon -lt -70 -and $lon -gt -72.5) {
    return @{ lat = $lat; long = $lon }
  }
  return $null
}

# Clean trims and HTML-decodes (Boston's export double-encodes entities, e.g.
# "&amp;" in addresses). Returns $null for empty/whitespace.
function Clean([string]$val) {
  if ([string]::IsNullOrWhiteSpace($val)) { return $null }
  return ([System.Net.WebUtility]::HtmlDecode($val)).Trim()
}

function Add-IfSet([hashtable]$h, [string]$key, [string]$val) {
  $c = Clean $val
  if ($c) { $h[$key] = $c }
}

function Get-Status([string]$s) {
  switch -Regex ($s) {
    'closed'  { return 'closed' }
    'open'    { return 'open' }
    default   { return 'open' }   # OVERDUE, On Time, etc. are still open cases
  }
}

# Map one CSV row to the Open311 request body, or $null to skip.
function ConvertTo-Body($r) {
  $srid = $r.case_enquiry_id
  if ([string]::IsNullOrWhiteSpace($srid)) { return $null }

  $svcCode = Clean $r.type
  if (-not $svcCode) { $svcCode = Clean $r.reason }
  if (-not $svcCode) { return $null }

  $coord = Get-ValidCoord $r.latitude $r.longitude
  $addr = Clean $r.location
  if (-not $coord -and -not $addr) { return $null }  # no location -> API would 400

  $body = [ordered]@{
    service_request_id = $srid.Trim()
    service_code       = $svcCode
    service_name       = $svcCode
    status             = (Get-Status $r.case_status)
  }
  $desc = Clean $r.case_title
  if ($desc) { $body.description = $desc }
  $agency = Clean $r.department
  if ($agency) { $body.agency_responsible = $agency }

  $reqDt = ConvertTo-UtcIso $r.open_dt
  if ($reqDt) { $body.requested_datetime = $reqDt }
  $updDt = ConvertTo-UtcIso $r.closed_dt        # preserved by upsert; absent -> server stamps now
  if ($updDt) { $body.updated_datetime = $updDt }

  if ($addr) { $body.address = $addr }
  Add-IfSet $body 'zipcode' $r.location_zipcode
  if ($coord) { $body.lat = $coord.lat; $body.long = $coord.long }
  if ($r.submitted_photo -match '^https?://') { $body.media_url = $r.submitted_photo.Trim() }

  $props = @{}
  Add-IfSet $props 'subject'                        $r.subject
  Add-IfSet $props 'reason'                          $r.reason
  Add-IfSet $props 'type'                            $r.type
  Add-IfSet $props 'queue'                           $r.queue
  Add-IfSet $props 'department'                      $r.department
  Add-IfSet $props 'on_time'                         $r.on_time
  Add-IfSet $props 'closure_reason'                  $r.closure_reason
  Add-IfSet $props 'sla_target_dt'                   $r.sla_target_dt
  Add-IfSet $props 'source'                          $r.source
  Add-IfSet $props 'fire_district'                   $r.fire_district
  Add-IfSet $props 'pwd_district'                    $r.pwd_district
  Add-IfSet $props 'city_council_district'           $r.city_council_district
  Add-IfSet $props 'police_district'                 $r.police_district
  Add-IfSet $props 'neighborhood'                    $r.neighborhood
  Add-IfSet $props 'neighborhood_services_district'  $r.neighborhood_services_district
  Add-IfSet $props 'ward'                            $r.ward
  Add-IfSet $props 'precinct'                        $r.precinct
  Add-IfSet $props 'location_street_name'            $r.location_street_name
  Add-IfSet $props 'geom_4326'                       $r.geom_4326   # WKB lineage of the original geometry
  if ($props.Count -gt 0) { $body.properties = $props }

  return $body
}

# Send one HTTP request with adaptive Retry-After backoff (for when rate limiting
# is enabled). Returns the response object; throws on non-429 HTTP errors.
function Invoke-WithBackoff([string]$uri, [string]$method, [byte[]]$bytes, [hashtable]$hdrs, [string]$label) {
  while ($true) {
    try {
      return Invoke-WebRequest -Uri $uri -Method $method -Headers $hdrs `
               -ContentType 'application/json' -Body $bytes -UseBasicParsing
    } catch {
      $code = $null
      if ($_.Exception.Response) { $code = [int]$_.Exception.Response.StatusCode }
      if ($code -eq 429) {
        $ra = 0
        try { $ra = [int]$_.Exception.Response.Headers['Retry-After'] } catch {}
        if ($ra -lt 1) { $ra = 60 }
        Write-Host ("  429 $label -> waiting {0}s" -f $ra) -ForegroundColor Yellow
        Start-Sleep -Seconds ($ra + 1)
      } else { throw }
    }
  }
}

# --- main ---
$rows = Import-Csv $CsvPath | Select-Object -Skip $Skip -First $First
Write-Host ("Selected {0} row(s) (Skip={1}, First={2}) from {3}" -f $rows.Count, $Skip, $First, $CsvPath)

$headers = @{ 'X-API-Key' = $ApiKey; 'Accept' = 'application/json' }
$created = 0; $updated = 0; $skipped = 0; $failed = 0; $i = 0

if ($Bulk) {
  # --- bulk mode: chunk mapped bodies and POST arrays to /requests/bulk ---
  $batch = New-Object System.Collections.ArrayList
  $batchNo = 0

  function Send-Batch {
    param($items)
    if ($items.Count -eq 0) { return }
    $script:batchNo++
    $json = ConvertTo-Json -InputObject @($items) -Depth 6 -Compress
    if ($DryRun) {
      Write-Host ("[batch {0}] DRY {1} record(s)" -f $script:batchNo, $items.Count) -ForegroundColor Cyan
      if ($script:batchNo -eq 1) { Write-Host $json }
      return
    }
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($json)
    $resp = Invoke-WithBackoff "$BaseUrl/requests/bulk" 'Post' $bytes $headers ("batch $script:batchNo")
    $r = $resp.Content | ConvertFrom-Json
    $script:created += [int]$r.created
    $script:updated += [int]$r.updated
    $script:failed  += [int]$r.failed
    Write-Host ("[batch {0}] {1} records -> created={2} updated={3} failed={4}" -f `
      $script:batchNo, $items.Count, $r.created, $r.updated, $r.failed) -ForegroundColor Green
  }

  foreach ($row in $rows) {
    $i++
    $body = ConvertTo-Body $row
    if (-not $body) { $skipped++; continue }
    [void]$batch.Add($body)
    if ($batch.Count -ge $BatchSize) { Send-Batch $batch; $batch.Clear() }
  }
  Send-Batch $batch
}
else {
  # --- per-record mode: one PUT per row (paced for rate limits) ---
  foreach ($row in $rows) {
    $i++
    $body = ConvertTo-Body $row
    if (-not $body) {
      $skipped++
      Write-Host ("[{0}] SKIP  {1} (no location or no service_code)" -f $i, $row.case_enquiry_id) -ForegroundColor DarkYellow
      continue
    }
    $srid = $body.service_request_id
    $json = $body | ConvertTo-Json -Depth 6 -Compress
    if ($DryRun) {
      Write-Host ("[{0}] DRY   {1}" -f $i, $srid) -ForegroundColor Cyan
      Write-Host $json
      continue
    }
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($json)
    try {
      $resp = Invoke-WithBackoff "$BaseUrl/requests/$srid" 'Put' $bytes $headers $srid
      if ($resp.StatusCode -eq 201) { $created++; $tag = 'CREATE 201' } else { $updated++; $tag = 'UPDATE 200' }
      Write-Host ("[{0}] {1} {2}" -f $i, $tag, $srid) -ForegroundColor Green
    } catch {
      $code = $null
      if ($_.Exception.Response) { $code = [int]$_.Exception.Response.StatusCode }
      $failed++
      Write-Host ("[{0}] FAIL  {1} (HTTP {2})" -f $i, $srid, $code) -ForegroundColor Red
    }
    Start-Sleep -Seconds $DelaySeconds
  }
}

Write-Host ""
Write-Host ("Done. created=$created updated=$updated skipped=$skipped failed=$failed selected=$($rows.Count)") -ForegroundColor White
