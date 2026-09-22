param(
    [Parameter(Mandatory = $true, ParameterSetName = "Feature")]
    [string]$Feature,
    [Parameter(Mandatory = $true, ParameterSetName = "Issue")]
    [int]$Issue,
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

$arguments = @("workflow", "run", "feature-plan.yml", "--ref", $Ref, "-f", "mode=$Mode")
if ($PSCmdlet.ParameterSetName -eq "Issue") {
    $arguments += @("-f", "issue=$Issue")
} else {
    $arguments += @("-f", "feature=$Feature")
}
gh @arguments
if ($LASTEXITCODE -ne 0) {
    throw "No se pudo iniciar el workflow feature-plan.yml."
}

if ($PSCmdlet.ParameterSetName -eq "Issue") {
    Write-Host "Workflow iniciado para el Issue #$Issue en modo '$Mode'."
} else {
    Write-Host "Workflow iniciado para la feature '$Feature' en modo '$Mode'."
}
Write-Host "Consulta la ejecución con: gh run list --workflow feature-plan.yml --limit 1"