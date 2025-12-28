// MongoDB initialization script for mongosh
// Usage:
//   mongosh "<CONNECTION_STRING>" --file init_mongo.js
// Example:
//   mongosh "mongodb+srv://admin:PASS@cluster0.mongodb.net/admin" --file init_mongo.js

const DB_NAME = "betting";
const DB_USER = "betting_user"; // change this
const DB_PASS = "CHANGE_ME";    // change this

const db = db.getSiblingDB(DB_NAME);

// Create collections if they don't exist
try {
  db.createCollection("users");
} catch (e) {}
try {
  db.createCollection("transactions");
} catch (e) {}
try {
  db.createCollection("games");
} catch (e) {}

// Create indexes used by the app
db.users.createIndex({ uid: 1 }, { unique: true });
db.transactions.createIndex({ userId: 1 });
db.games.createIndex({ createdAt: -1 });

// Create a dedicated DB user (readWrite on the DB)
// If connecting to Atlas, create users via the Atlas UI instead for best practice.
try {
  db.createUser({
    user: DB_USER,
    pwd: DB_PASS,
    roles: [{ role: "readWrite", db: DB_NAME }]
  });
  print(`Created DB user '${DB_USER}' on '${DB_NAME}'.`);
} catch (e) {
  print(`Could not create DB user: ${e}`);
}

print(`Initialization complete for DB: ${DB_NAME}`);
