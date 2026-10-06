# 📦 Package `database`

**Source Path:** `pkg/database`

## 🚀 Functions

### `NewMongoDatabase`

NewMongoDatabase opens the MongoDB connection used by every service and
makes sure the collection indexes exist.

It takes a logger.LoggerInterface as argument and returns a
*mongo.Database and an error. The logger is used to log any errors that
occur when connecting to the database.

The connection settings are resolved by `pkg/database/mongo` (via viper)
from any of:

- MONGO_URI or DB_URL: a full MongoDB connection string (takes precedence)
- MONGO_HOST / DB_HOST: the database host
- MONGO_PORT / DB_PORT: the database port (default 27017)
- MONGO_USERNAME / DB_USERNAME: the username
- MONGO_PASSWORD / DB_PASSWORD: the password
- MONGO_DATABASE / DB_NAME: the database name
- MONGO_MAX_POOL_SIZE / MONGO_MIN_POOL_SIZE: connection pool sizing

The function will return an error if the connection or index creation
fails.

```go
func NewMongoDatabase(logger logger.LoggerInterface) (*mongo.Database, error)
```
