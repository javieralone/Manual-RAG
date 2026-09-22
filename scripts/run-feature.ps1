param(
    [Parameter(Mandatory = $true)]
    [string]$Feature,
    [ValidateSet("plan", "implement")]
    [string]$Mode = "plan",
    [string]$Ref = "dev"
)

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    throw "GitHub CLI (gh) no está instalado o no está en PATH."
}

gh auth status
if ($LASTEXITCODE -ne 0) {
    throw "Autentica GitHub CLI con 'gh auth login' antes de ejecutar este comando."
}

gh workflow run feature-plan.yml --ref $Ref -f feature=$Feature -f mode=$Mode
if ($LASTEXITCODE -ne 0) {
    throw "No se pudo iniciar el workflow feature-plan.yml."
}

Write-Host "Workflow iniciado para la feature '$Feature' en modo '$Mode'."
Write-Host "Consulta la ejecución con: gh run list --workflow feature-plan.yml --limit 1"