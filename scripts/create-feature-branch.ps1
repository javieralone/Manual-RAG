param(
    [Parameter(Mandatory = $true)]
    [string]$Name
)

$slug = $Name.ToLowerInvariant()
$slug = [regex]::Replace($slug, '[^a-z0-9-]+', '-')
$slug = [regex]::Replace($slug, '-+', '-')
$slug = $slug.Trim('-')

if ([string]::IsNullOrWhiteSpace($slug)) {
    throw "El nombre de la feature no es válido."
}

$branch = "feature/$slug"
git checkout -b $branch
Write-Host "Rama creada: $branch"
