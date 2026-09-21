param(
    [string]$Feature,
    [switch]$Push
)

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not $Feature) {
    $Feature = "01-base-observabilidad"
}

$targetBranch = "feature/$Feature"

$branchExists = git rev-parse --verify $targetBranch 2>$null
if (-not $branchExists) {
    git checkout -b $targetBranch
} else {
    git checkout $targetBranch
}

& "$PSScriptRoot\validate-pr.sh"
if ($LASTEXITCODE -ne 0) {
    throw "La validación falló. No se continúa con la siguiente feature."
}

if ($Push) {
    git add -A
    git diff --cached --quiet
    if ($LASTEXITCODE -ne 0) {
        git commit -m "feat: $Feature"
    }
    git push --set-upstream origin $targetBranch
}

Write-Host "Feature '$Feature' validada y lista para continuar."
