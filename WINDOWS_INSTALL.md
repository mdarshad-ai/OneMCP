# Windows Installation Instructions

## Method 1: Direct Download & Install
1. Download onemcp.exe from GitHub Releases
2. Create folder: C:\Program Files\OneMCP\
3. Move onemcp.exe to that folder
4. Add to PATH: System Properties → Environment Variables → Path → Add 'C:\Program Files\OneMCP\'

## Method 2: User Directory (Recommended)
1. Download onemcp.exe
2. Create folder: %USERPROFILE%\bin\
3. Move onemcp.exe there
4. Add %USERPROFILE%\bin\ to PATH

## Method 3: PowerShell Installation
Invoke-WebRequest -Uri 'https://github.com/mdarshad-ai/OneMCP/releases/download/v1.0.0/onemcp.exe' -OutFile '$env:USERPROFILE\bin\onemcp.exe'

