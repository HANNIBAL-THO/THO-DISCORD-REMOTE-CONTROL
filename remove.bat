@echo off
echo ==========================================
echo Eliminando el programa del arranque...
echo ==========================================

reg delete "HKCU\Software\Microsoft\Windows\CurrentVersion\Run" /v DiscordRemoteControl /f >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] Entrada de inicio eliminada correctamente.
) else (
    echo [ERROR] No se encontró la entrada de inicio o no se pudo eliminar.
)

echo.
echo ==========================================
echo Eliminando el ejecutable...
echo ==========================================

set EXE_PATH=%~dp0THO-DISCORD-CONTROL-REMOTE.exe
if exist "%EXE_PATH%" (
    del /f /q "%EXE_PATH%"
    if %ERRORLEVEL% EQU 0 (
        echo [OK] Ejecutable eliminado correctamente.
    ) else (
        echo [ERROR] No se pudo eliminar el ejecutable.
    )
) else (
    echo [INFO] El ejecutable no se encontró.
)

echo.
echo ==========================================
echo Eliminando archivos temporales...
echo ==========================================

set TEMP_FILES=%~dp0*.tmp
if exist "%TEMP_FILES%" (
    del /f /q "%TEMP_FILES%"
    if %ERRORLEVEL% EQU 0 (
        echo [OK] Archivos temporales eliminados correctamente.
    ) else (
        echo [ERROR] No se pudieron eliminar los archivos temporales.
    )
) else (
    echo [INFO] No se encontraron archivos temporales.
)

echo.
echo ==========================================
echo Operación completada.
echo ==========================================
pause