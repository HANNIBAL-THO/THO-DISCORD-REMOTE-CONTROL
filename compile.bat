@echo off
echo Compilando tho Discord Remote Control...

REM 
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

REM 
if exist "build\" rmdir /s /q "build"
mkdir build

REM 
echo Actualizando dependencias...
go mod tidy

REM 
echo Compilando con optimizaciones de tamaño...
go build -trimpath -ldflags="-s -w -H=windowsgui" -o build/THO-DISCORD-CONTROL-REMOTE.exe

REM 
echo.
echo Compilación completada! Los archivos están en la carpeta 'build'
echo Tamaño del ejecutable final:
dir build\THO-DISCORD-CONTROL-REMOTE.exe | findstr /C:"THO-DISCORD-CONTROL-REMOTE.exe"
pause

