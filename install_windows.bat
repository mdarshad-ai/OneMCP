@echo off
echo OneMCP Windows Installer
echo =======================

echo Creating bin directory...
if not exist "%USERPROFILE%\bin" mkdir "%USERPROFILE%\bin"

echo Downloading OneMCP...
powershell -Command "Invoke-WebRequest -Uri 'https://github.com/mdarshad-ai/OneMCP/releases/download/v1.0.0/onemcp.exe' -OutFile '%USERPROFILE%\bin\onemcp.exe'"

echo Adding to PATH...
setx PATH "%PATH%;%USERPROFILE%\bin"

echo Installation complete!
echo Run 'onemcp --help' to get started
echo You may need to restart your command prompt for PATH changes to take effect.

pause
