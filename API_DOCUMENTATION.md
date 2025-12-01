# Backend API Documentation

## Base URL
```
http://localhost:4001
```

## Authentication
All protected endpoints require a Firebase ID token in the Authorization header:
```
Authorization: Bearer <firebase_id_token>
```

---

## Health Check

### GET /api/health
Check backend service status.

**Response:**
```json
{
  "status": "running",
  "firebaseReady": true,
  "mongodbReady": true,
  "timestamp": 1701234567
}
```

---

## Authentication Endpoints

### POST /api/auth/verify
Verify Firebase ID token and create/update user in database.

**Request:**
```json
{
  "idToken": "eyJhbGc..."
}
```

**Response:**
```json
{
  "uid": "firebase_user_id",
  "email": "user@example.com",
  "name": "User Name"
}
```

---

## Wallet Endpoints (Protected)

### GET /api/wallet/balance
Get user's wallet balance.

**Headers:** Authorization required

**Response:**
```json
{
  "userId": "user_id",
  "balance": 5000.00,
  "currency": "INR",
  "lockedBalance": 0,
  "lastUpdated": "2025-11-30T12:00:00Z",
  "createdAt": "2025-11-01T10:00:00Z"
}
```

### POST /api/wallet/payment-request
Create a new deposit request.

**Headers:** Authorization required

**Request:**
```json
{
  "amount": 1000,
  "paymentMethod": "upi",
  "transactionId": "TXN123456",
  "proofUrl": "https://...",
  "notes": "Payment from PhonePe"
}
```

**Response:**
```json
{
  "id": "req_1234567890",
  "userId": "user_id",
  "amount": 1000,
  "paymentMethod": "upi",
  "status": "pending",
  "createdAt": "2025-11-30T12:00:00Z",
  "updatedAt": "2025-11-30T12:00:00Z"
}
```

### GET /api/wallet/payment-requests
Get user's payment requests (or all if admin).

**Headers:** Authorization required

**Query Parameters:**
- `status` (optional): Filter by status (pending, accepted, declined, all)

**Response:**
```json
[
  {
    "id": "req_1234567890",
    "userId": "user_id",
    "amount": 1000,
    "paymentMethod": "upi",
    "status": "pending",
    "createdAt": "2025-11-30T12:00:00Z"
  }
]
```

### GET /api/wallet/transactions
Get user's transaction history.

**Headers:** Authorization required

**Query Parameters:**
- `limit` (optional): Number of transactions to return (default: 50)

**Response:**
```json
[
  {
    "id": "txn_1234567890",
    "userId": "user_id",
    "type": "credit",
    "amount": 1000,
    "description": "Payment request accepted",
    "category": "deposit",
    "balanceBefore": 4000,
    "balanceAfter": 5000,
    "status": "completed",
    "createdAt": "2025-11-30T12:00:00Z"
  }
]
```

### GET /api/wallet/payment-details
Get admin payment details for deposits.

**Response:**
```json
{
  "bankName": "HDFC Bank",
  "accountNumber": "1234567890",
  "ifscCode": "HDFC0001234",
  "accountHolderName": "NeonPlay Gaming Pvt Ltd",
  "upiId": "neonplay@hdfc",
  "qrCodeUrl": "/payment-qr.png",
  "updatedAt": "2025-11-30T12:00:00Z"
}
```

---

## Game Endpoints (Protected)

### POST /api/game/play
Record a game play result.

**Headers:** Authorization required

**Request:**
```json
{
  "gameType": "spinwheel",
  "betAmount": 100,
  "winAmount": 500,
  "multiplier": 5,
  "resultData": {
    "segment": 4,
    "prize": "5x Win!"
  }
}
```

**Response:**
```json
{
  "id": "game_1234567890",
  "userId": "user_id",
  "gameType": "spinwheel",
  "betAmount": 100,
  "winAmount": 500,
  "multiplier": 5,
  "settled": true,
  "createdAt": "2025-11-30T12:00:00Z"
}
```

**Error Responses:**
- `402 Payment Required`: Insufficient balance
- `400 Bad Request`: Invalid game data

### GET /api/game/history
Get user's game history.

**Headers:** Authorization required

**Query Parameters:**
- `game_type` (optional): Filter by game type (spinwheel, slot, etc.)
- `limit` (optional): Number of games to return (default: 50)

**Response:**
```json
[
  {
    "id": "game_1234567890",
    "userId": "user_id",
    "gameType": "spinwheel",
    "betAmount": 100,
    "winAmount": 500,
    "createdAt": "2025-11-30T12:00:00Z"
  }
]
```

### GET /api/game/stats
Get user's game statistics.

**Headers:** Authorization required

**Query Parameters:**
- `game_type` (optional): Filter by game type

**Response:**
```json
{
  "total_games": 150,
  "total_wagered": 15000,
  "total_won": 12000,
  "biggest_win": 5000,
  "by_game_type": [
    {
      "_id": "spinwheel",
      "total_games": 50,
      "total_wagered": 5000,
      "total_won": 4000
    }
  ]
}
```

### GET /api/game/recent-bets
Get recent bets across all users (public).

**Query Parameters:**
- `limit` (optional): Number of bets to return (default: 20)

**Response:**
```json
[
  {
    "id": "game_1234567890",
    "gameType": "spinwheel",
    "betAmount": 100,
    "winAmount": 500,
    "createdAt": "2025-11-30T12:00:00Z"
  }
]
```

---

## Admin Endpoints (Protected - Admin Only)

### GET /api/admin/payment-requests
Get all payment requests (admin only).

**Headers:** Authorization required (admin role)

**Query Parameters:**
- `status` (optional): Filter by status

**Response:** Same as user payment requests but includes all users

### POST /api/admin/payment-request/:id/approve
Approve a payment request (credits user wallet).

**Headers:** Authorization required (admin role)

**Request:**
```json
{
  "adminNotes": "Verified payment screenshot"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Payment request accepted"
}
```

### POST /api/admin/payment-request/:id/decline
Decline a payment request.

**Headers:** Authorization required (admin role)

**Request:**
```json
{
  "adminNotes": "Invalid transaction ID"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Payment request declined"
}
```

### PUT /api/admin/payment-details
Update admin payment details.

**Headers:** Authorization required (admin role)

**Request:**
```json
{
  "bankName": "HDFC Bank",
  "accountNumber": "1234567890",
  "ifscCode": "HDFC0001234",
  "accountHolderName": "NeonPlay Gaming Pvt Ltd",
  "upiId": "neonplay@hdfc",
  "qrCodeUrl": "/payment-qr.png"
}
```

**Response:** Updated payment details object

---

## User Profile Endpoints (Protected)

### GET /api/user/profile
Get user profile information.

**Headers:** Authorization required

**Response:**
```json
{
  "uid": "user_id",
  "email": "user@example.com",
  "name": "User Name",
  "phone": "+919876543210",
  "role": "user",
  "totalGamesPlayed": 150,
  "totalWagered": 15000,
  "totalWon": 12000,
  "createdAt": "2025-11-01T10:00:00Z"
}
```

### PUT /api/user/profile
Update user profile.

**Headers:** Authorization required

**Request:**
```json
{
  "name": "New Name",
  "profilePic": "https://..."
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Profile updated successfully"
}
```

### GET /api/user/stats
Get user statistics summary.

**Headers:** Authorization required

**Response:**
```json
{
  "totalGamesPlayed": 150,
  "totalWagered": 15000,
  "totalWon": 12000
}
```

---

## MongoDB Collections Schema

### users
```javascript
{
  uid: String (unique),
  email: String,
  name: String,
  phone: String,
  role: String,
  total_games_played: Number,
  total_wagered: Number,
  total_won: Number,
  created_at: Date,
  last_seen_at: Date
}
```

### wallets
```javascript
{
  user_id: String (unique),
  balance: Number,
  currency: String,
  locked_balance: Number,
  last_updated: Date,
  created_at: Date
}
```

### transactions
```javascript
{
  _id: String,
  user_id: String,
  type: String, // credit, debit
  amount: Number,
  description: String,
  category: String, // deposit, withdrawal, game_win, game_loss
  balance_before: Number,
  balance_after: Number,
  status: String,
  created_at: Date
}
```

### games
```javascript
{
  _id: String,
  user_id: String,
  game_type: String,
  bet_amount: Number,
  win_amount: Number,
  multiplier: Number,
  result_data: Object,
  settled: Boolean,
  created_at: Date
}
```

### payment_requests
```javascript
{
  _id: String,
  user_id: String,
  amount: Number,
  payment_method: String,
  transaction_id: String,
  proof_url: String,
  status: String, // pending, accepted, declined
  notes: String,
  admin_notes: String,
  created_at: Date,
  updated_at: Date
}
```

### payment_details
```javascript
{
  bank_name: String,
  account_number: String,
  ifsc_code: String,
  account_holder_name: String,
  upi_id: String,
  qr_code_url: String,
  updated_at: Date
}
```

---

## Game Types

Supported game types:
- `aviation` - Aviation crash game
- `spinwheel` - Spin the wheel
- `slot` - Slot machine
- `mines` - Mines game
- `plinko` - Plinko game
- `dice` - Dice roll
- `limbo` - Limbo game
- `hilo` - Hi-Lo card game
- `blackjack` - Blackjack card game

---

## Error Handling

All endpoints return appropriate HTTP status codes:
- `200 OK` - Successful request
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Missing or invalid authentication
- `402 Payment Required` - Insufficient balance
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

Error response format:
```json
{
  "error": "Error message description"
}
```
