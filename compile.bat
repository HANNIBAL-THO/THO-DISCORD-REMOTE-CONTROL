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
where upx >nul 2>nul
if %ERRORLEVEL% EQU 0 (
    echo Comprimiendo ejecutable con UPX...
    upx --best --lzma build/THO-DISCORD-CONTROL-REMOTE.exe
) else (
    echo UPX no encontrado. Para reducir aún más el tamaño, instala UPX desde https://github.com/upx/upx/releases
    echo y agrégalo al PATH del sistema.
)

REM 
copy config.json build\
echo.
echo Compilación completada! Los archivos están en la carpeta 'build'
echo Tamaño del ejecutable final:
dir build\THO-DISCORD-CONTROL-REMOTE.exe | findstr /C:"THO-DISCORD-CONTROL-REMOTE.exe"
pause
