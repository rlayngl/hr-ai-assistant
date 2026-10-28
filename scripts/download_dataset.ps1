$ErrorActionPreference = 'Stop'
$url = 'https://raw.githubusercontent.com/aniketnandi/automated-resume-screening/main/data/UpdatedResumeDataSet.csv'
$out = Join-Path $PSScriptRoot '..\data\UpdatedResumeDataSet.csv'
$out = [System.IO.Path]::GetFullPath($out)
Invoke-WebRequest -Uri $url -OutFile $out
Write-Host "Downloaded public resume dataset to $out"
