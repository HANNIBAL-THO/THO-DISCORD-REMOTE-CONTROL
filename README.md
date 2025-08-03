# THO-DISCORD-CONTROL-REMOTO

## 🚨 Descargo de Responsabilidad

Este software ha sido creado con fines educativos y de pruebas personales. El usuario es el único responsable del uso que le dé a esta herramienta. El uso indebido de este software puede violar leyes de privacidad y seguridad informática.

## 📋 Descripción

THO-DISCORD-CONTROL-REMOTO es una herramienta de control remoto que permite administrar y monitorear sistemas a través de un bot Discord. Permite ejecutar diversos comandos para obtener información del sistema, capturar pantallas, deshabilitar conexiones de red, y más.

## ✨ Características

- 💻 Información detallada del sistema
- 📸 Captura de pantalla remota
- 🌐 Información de red y geolocalización
- 📶 Desabilitacion de conexiones WiFi
- 🔒 Bloqueo remoto del sistema
- 🔊 Envío de alertas sonoras
- 🌍 Apertura de URLs en el navegador
- 🔄 Reinicio y apagado remoto
- 🔑 Recuperación de credenciales (solo con fines educativos)

## 🛠️ Instalación

1. Clona este repositorio:
```bash
git clone https://github.com/tu-usuario/THO-DISCORD-CONTROL-REMOTO.git
cd THO-DISCORD-CONTROL-REMOTO
```

2. Configura tu token del bot de Discord en un archivo `config.json`:
```json
{
  "token": "TU_TOKEN_DE_DISCORD",
  "prefix": "!",
  "owner_id": "TU_ID_DE_DISCORD",
  "guild_id": "ID_DEL_SERVIDOR"
}
```

3. Compila el proyecto:
```bash
go build -o THO-DISCORD-CONTROL-REMOTE.exe
```

Alternativamente, puedes usar el script `compile.bat` incluido.

## 🚀 Uso

Ejecuta el archivo compilado:
```bash
THO-DISCORD-CONTROL-REMOTE.exe
```

Una vez iniciado, el bot se conectará a Discord y estará listo para recibir comandos.

### Comandos Disponibles

- `!help` - Muestra la lista de comandos disponibles
- `!system` - Muestra información completa del sistema
- `!status` - Muestra el estado del sistema
- `!screenshot` - Toma una captura de pantalla y la envía
- `!slowwifi` - Deshabilita todas las conexiones de red
- `!location` - Muestra la localización aproximada por IP
- `!ipinfo` - Muestra la IP pública y local
- `!blockall` - Bloquea el sistema completamente
- `!alert` - Envía una alerta sonora con el mensaje indicado
- `!openurl` - Abre una URL en el navegador predeterminado
- `!reboot` - Reinicia el sistema
- `!shutdown` - Apaga el sistema
- `!credentials` - Obtiene credenciales y datos sensibles del sistema
- `!startup` - Lista los programas de inicio
- `!updates` - Verifica actualizaciones pendientes
- `!remove` - Elimina el ejecutable y lo quita del inicio

## 🔧 Requisitos

- Go 1.16 o superior
- upx 5.0.1
- Sistema operativo Windows (algunas funciones también son compatibles con Linux)
- Token de bot de Discord

## 📝 Notas

- Este es mi discord : https://discord.gg/uPESr5v7yQ
- Este software debe ser utilizado únicamente en sistemas propios o con autorización explícita.
- No me hago responsable del mal uso de esta herramienta.
- Este proyecto fue creado con fines educativos para entender cómo funcionan las herramientas de administración remota.

## 📜 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo LICENSE para más detalles.

---


⚠️ ADVERTENCIA: El uso de este software para acceder a sistemas sin autorización puede constituir un delito. Úselo bajo su propia responsabilidad y solo en entornos controlados con fines educativos.



