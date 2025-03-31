const express = require('express');
const cors = require('cors');
const bodyParser = require('body-parser');

const app = express();
const PORT = 3001;

// CORS設定
app.use(cors());
app.use(bodyParser.json());

// ダミーの認証レスポンス - 実際のプロジェクトでは実際のgRPCクライアントを実装する
app.post('/auth/register', (req, res) => {
  console.log('Register request received:', req.body);
  
  // ダミーレスポンス
  res.json({
    user: {
      id: '123e4567-e89b-12d3-a456-426614174000',
      email: req.body.email,
      firstName: req.body.firstName,
      lastName: req.body.lastName,
      phoneNumber: req.body.phoneNumber || null
    },
    token: 'dummy-jwt-token-for-testing-purposes-only'
  });
});

app.post('/auth/login', (req, res) => {
  console.log('Login request received:', req.body);
  
  // ダミーレスポンス
  res.json({
    user: {
      id: '123e4567-e89b-12d3-a456-426614174000',
      email: req.body.email,
      firstName: 'テスト',
      lastName: 'ユーザー',
      phoneNumber: null
    },
    token: 'dummy-jwt-token-for-testing-purposes-only'
  });
});

app.get('/users/profile', (req, res) => {
  // ここで認証トークンをチェックする必要があります
  res.json({
    id: '123e4567-e89b-12d3-a456-426614174000',
    email: 'test@example.com',
    firstName: 'テスト',
    lastName: 'ユーザー',
    phoneNumber: null
  });
});

app.put('/users/profile', (req, res) => {
  console.log('Profile update request received:', req.body);
  
  // ダミーレスポンス - 更新されたデータを返す
  res.json({
    id: '123e4567-e89b-12d3-a456-426614174000',
    email: 'test@example.com',
    firstName: req.body.firstName || 'テスト',
    lastName: req.body.lastName || 'ユーザー',
    phoneNumber: req.body.phoneNumber || null
  });
});

// 住所関連のエンドポイント
let addresses = []; // シンプルなインメモリストレージ

app.get('/users/addresses', (req, res) => {
  res.json(addresses);
});

app.post('/users/addresses', (req, res) => {
  const newAddress = {
    id: `addr-${Date.now()}`,
    ...req.body
  };
  
  addresses.push(newAddress);
  res.status(201).json(newAddress);
});

app.put('/users/addresses/:id', (req, res) => {
  const id = req.params.id;
  const index = addresses.findIndex(addr => addr.id === id);
  
  if (index === -1) {
    return res.status(404).json({ message: '住所が見つかりません' });
  }
  
  addresses[index] = {
    ...addresses[index],
    ...req.body
  };
  
  res.json(addresses[index]);
});

app.delete('/users/addresses/:id', (req, res) => {
  const id = req.params.id;
  addresses = addresses.filter(addr => addr.id !== id);
  res.status(204).send();
});

// サーバー起動
app.listen(PORT, () => {
  console.log(`API Proxy Server is running on http://localhost:${PORT}`);
}); 