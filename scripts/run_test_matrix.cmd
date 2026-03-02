@echo off
setlocal ENABLEDELAYEDEXPANSION

set ROOT_DIR=%~dp0..
set LOG_DIR=%ROOT_DIR%\tests\logs
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"

echo [1/4] go unit+integration tests
pushd "%ROOT_DIR%"
go test ./...
if errorlevel 1 goto :fail
popd

echo [2/4] python unit tests
python3 -m unittest discover -s "%ROOT_DIR%\ml\inference" -p test_*.py
if errorlevel 1 goto :fail

echo [3/4] sidecar + runtime smoke
set SOCK=%LOG_DIR%\blackice-matrix.sock
if exist "%SOCK%" del /F /Q "%SOCK%"
set PY_LOG=%LOG_DIR%\python_smoke.log
set GO_LOG=%LOG_DIR%\go_smoke.log

start "blackice-py" /B cmd /C "python3 "%ROOT_DIR%\ml\inference\server.py" --socket %SOCK% > "%PY_LOG%" 2>&1"
timeout /T 1 /NOBREAK >NUL
start "blackice-go" /B cmd /C "cd /D "%ROOT_DIR%" && go run ./cmd/blackice --socket %SOCK% --pps 120000 --window 1s > "%GO_LOG%" 2>&1"
timeout /T 3 /NOBREAK >NUL

taskkill /FI "WINDOWTITLE eq blackice-go" /T /F >NUL 2>&1
taskkill /FI "WINDOWTITLE eq blackice-py" /T /F >NUL 2>&1

echo [4/4] quick assertion of mitigation output
findstr /C:"mitigation=" "%GO_LOG%" >NUL
if errorlevel 1 goto :missing

echo matrix complete
goto :eof

:missing
echo expected mitigation output missing
exit /b 1

:fail
exit /b 1
