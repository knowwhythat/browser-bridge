@echo off
chcp 65001 >nul 2>&1
setlocal

echo ============================================
echo  Browser Bridge - Installation Script
echo ============================================
echo.

set "SCRIPT_DIR=%~dp0"
set "ROOT_DIR=%SCRIPT_DIR%.."

for %%i in ("%ROOT_DIR%") do set "ROOT_DIR=%%~fi"

echo Project root: %ROOT_DIR%
echo.

echo [1/4] Building native-host...
cd /d "%ROOT_DIR%\native-host"
go build -o native-host.exe .
if errorlevel 1 (
    echo ERROR: Failed to build native-host
    pause
    exit /b 1
)
echo   native-host.exe built successfully

echo [2/4] Building browser-bridge CLI...
cd /d "%ROOT_DIR%\cli"
go build -o browser-bridge.exe .
if errorlevel 1 (
    echo ERROR: Failed to build CLI
    pause
    exit /b 1
)
echo   browser-bridge.exe built successfully

echo [3/4] Building Chrome Extension...
cd /d "%ROOT_DIR%\extension"
if not exist "node_modules" (
    echo   Installing dependencies...
    call npm install
    if errorlevel 1 (
        echo ERROR: Failed to install extension dependencies
        pause
        exit /b 1
    )
)
call npm run build
if errorlevel 1 (
    echo ERROR: Failed to build extension
    pause
    exit /b 1
)
echo   Extension built successfully

echo [4/4] Registering Native Messaging Host...
set "NATIVE_HOST_EXE=%ROOT_DIR%\native-host\native-host.exe"
set "JSON_CONFIG=%ROOT_DIR%\install\com.browser.bridge.json"

powershell -Command "$j=Get-Content '%JSON_CONFIG%' -Raw | ConvertFrom-Json; $j.path='%NATIVE_HOST_EXE%'; $j | ConvertTo-Json | Set-Content '%JSON_CONFIG%' -Encoding UTF8"

reg add "HKCU\Software\Google\Chrome\NativeMessagingHosts\com.browser.bridge" /ve /t REG_SZ /d "%JSON_CONFIG%" /f
if errorlevel 1 (
    echo ERROR: Failed to write registry
    pause
    exit /b 1
)
echo   Native Messaging Host registered

echo.
echo ============================================
echo  Installation complete!
echo ============================================
echo.
echo  Next steps:
echo  1. Open Chrome and go to chrome://extensions/
echo  2. Enable Developer mode
echo  3. Click Load unpacked and select: %ROOT_DIR%\extension\dist
echo  4. Restart Chrome
echo.
pause
