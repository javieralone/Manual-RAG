param(
    [Parameter(Mandatory = $true)]
    [int]$Issue,
    [string]$Ref = "dev",
    [switch]$Wait
)

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

& "$PSScriptRoot\run-feature.ps1" -Issue $Issue -Mode implement -Ref $Ref
if ($LASTEXITCODE -ne 0) {
    throw "No se pudo iniciar la implementación del Issue #$Issue."
}

if ($Wait) {
    Write-Host "Esperando la ejecución de GitHub Actions..."
    $runId = gh run list --workflow feature-plan.yml --limit 1 --json databaseId --jq '.[0].databaseId'
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($runId)) {
        throw "No se pudo localizar la ejecución recién iniciada."
    }
    gh run watch $runId --exit-status
    if ($LASTEXITCODE -ne 0) {
        throw "El workflow falló. Revisa la ejecución $runId en GitHub."
    }
    Write-Host "Plan generado y publicado en el Issue #$Issue."
} else {
    Write-Host "El workflow leerá el Issue #$Issue, generará el plan y lo comentará allí."
    Write-Host "Usa '-Wait' para esperar a que termine la ejecución de GitHub Actions."
}