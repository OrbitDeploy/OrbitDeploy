@echo off

:: This script compiles the OrbitDeploy application for Linux (amd64) on Windows.

:: --- Configuration ---
set BINARY_NAME=orbit-deploy

:: --- Frontend Build ---
echo Building frontend...
cd frontend
where bun >nul 2>nul
if %errorlevel% neq 0 (
  echo bun is not installed. Please install it first.
  exit /b 1
)
bun install
bun run build
cd ..

:: --- Backend Build ---
echo Building OrbitDeploy for Linux (amd64)...

:: Set the target OS and architecture
set GOOS=linux
set GOARCH=amd64

:: Build the application
go build -o %BINARY_NAME% main.go

:: --- Post-build ---
if %errorlevel% equ 0 (
  echo.
  echo Build successful!
  echo The binary '%BINARY_NAME%' has been created in the current directory.
  echo You can now transfer this binary to your Linux server and run the install.sh script.
) else (
  echo.
  echo Build failed. Please check the error messages above.
  exit /b 1
)
