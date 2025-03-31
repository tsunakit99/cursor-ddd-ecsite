import {
  Box,
  Button,
  Card,
  CardContent,
  CardMedia,
  Grid,
  Typography
} from '@mui/material';
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';

const HomePage: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();

  // サンプル商品
  const products = [
    {
      id: '1',
      name: 'テスト商品1',
      description: 'これはテスト商品1の説明です。',
      price: 2000,
      imageUrl: 'https://placehold.co/300x200',
    },
    {
      id: '2',
      name: 'テスト商品2',
      description: 'これはテスト商品2の説明です。',
      price: 3500,
      imageUrl: 'https://placehold.co/300x200',
    },
    {
      id: '3',
      name: 'テスト商品3',
      description: 'これはテスト商品3の説明です。',
      price: 5000,
      imageUrl: 'https://placehold.co/300x200',
    },
    {
      id: '4',
      name: 'テスト商品4',
      description: 'これはテスト商品4の説明です。',
      price: 8000,
      imageUrl: 'https://placehold.co/300x200',
    },
  ];

  return (
    <Box>
      <Box sx={{ mb: 4, textAlign: 'center' }}>
        <Typography variant="h4" component="h1" gutterBottom>
          ECサイトへようこそ
        </Typography>
        <Typography variant="subtitle1" color="text.secondary" sx={{ mb: 2 }}>
          高品質な商品を豊富に取り揃えております
        </Typography>
        {!isAuthenticated && (
          <Box sx={{ mt: 2 }}>
            <Button 
              variant="contained" 
              color="primary" 
              size="large"
              onClick={() => navigate('/register')}
              sx={{ mr: 2 }}
            >
              新規登録
            </Button>
            <Button 
              variant="outlined" 
              color="primary" 
              size="large"
              onClick={() => navigate('/login')}
            >
              ログイン
            </Button>
          </Box>
        )}
      </Box>

      <Typography variant="h5" component="h2" sx={{ mb: 3 }}>
        おすすめ商品
      </Typography>
      <Grid container spacing={3}>
        {products.map((product) => (
          <Grid item xs={12} sm={6} md={3} key={product.id}>
            <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <CardMedia
                component="img"
                height="140"
                image={product.imageUrl}
                alt={product.name}
              />
              <CardContent sx={{ flexGrow: 1 }}>
                <Typography gutterBottom variant="h6" component="div">
                  {product.name}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  {product.description}
                </Typography>
                <Typography variant="h6" color="primary">
                  ¥{product.price.toLocaleString()}
                </Typography>
              </CardContent>
              <Box sx={{ p: 2, pt: 0 }}>
                <Button size="small" variant="contained" fullWidth>
                  カートに追加
                </Button>
              </Box>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Box>
  );
};

export default HomePage; 