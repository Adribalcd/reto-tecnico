# Demo en Azure

[Abrir la aplicación](https://reto-tecnico.gentlemoss-a8725290.eastus.azurecontainerapps.io)

Desplegada en **Azure Container Apps (Consumption)** con tres contenedores:
nginx sirve el frontend y redirige las peticiones a las APIs de Go y Node.js.
Solo el frontend se expone por HTTPS; las APIs se comunican internamente.

- Autenticación JWT con secretos almacenados en Azure. Las credenciales de acceso se entregan por separado.
- Escalado de cero a una réplica. La primera petición puede tardar mientras arranca la aplicación.
- Validado: login, endpoints protegidos, factorización QR y estadísticas; pruebas de Go y 23 pruebas de Node aprobadas.

La demo utiliza crédito de prueba con límite de gasto activado. El registro de imágenes consume crédito y la disponibilidad depende de la vigencia de la prueba.
