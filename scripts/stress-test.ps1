$ErrorActionPreference = "Stop"
$BaseUrl = $args[0]
if (-not $BaseUrl) { $BaseUrl = "https://chain.fycbit.com" }

$results = foreach ($i in 1..120) {
  try {
    $r = Invoke-WebRequest -UseBasicParsing -Uri "$BaseUrl/health" -TimeoutSec 5
    [string]$r.StatusCode
  } catch {
    if ($_.Exception.Response) {
      [string][int]$_.Exception.Response.StatusCode
    } else {
      "ERR"
    }
  }
}

$results | Group-Object | Sort-Object Name | ForEach-Object { "$($_.Name) $($_.Count)" }
