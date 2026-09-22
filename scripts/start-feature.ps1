param(
    [Parameter(Mandatory = $true)]
    [string]$Feature,
    [string]$Ref = "dev",
    [switch]$Wait
)

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

& "$PSScriptRoot\run-feature.ps1" -Feature $Feature -Mode implement -Ref $Ref
if ($LASTEXITCODE -ne 0) {
    throw "No se pudo iniciar la implementación de la feature '$Feature'."
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
    Write-Host "Plan generado e Issue de implementación creada correctamente."
} else {
    Write-Host "El workflow generará el plan y creará la Issue para Copilot."
    Write-Host "Usa '-Wait' para esperar a que termine la ejecución de GitHub Actions."
}