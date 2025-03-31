import {
    Alert,
    Box,
    Button,
    CircularProgress,
    Container,
    Link,
    Paper,
    TextField,
    Typography
} from '@mui/material';
import React, { useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useMutation } from 'react-query';
import { Link as RouterLink, useNavigate } from 'react-router-dom';
import { AuthApi, LoginRequest } from '../services/api';
import { useAuthStore } from '../store/authStore';

const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const { login } = useAuthStore();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const { control, handleSubmit, formState: { errors } } = useForm<LoginRequest>({
    defaultValues: {
      email: '',
      password: ''
    }
  });

  const loginMutation = useMutation(AuthApi.login, {
    onSuccess: (data) => {
      login(data.user, data.token);
      navigate('/');
    },
    onError: (error: any) => {
      console.error('Login error:', error);
      setErrorMessage(
        error.response?.data?.message || 
        'ログインに失敗しました。メールアドレスとパスワードを確認してください。'
      );
    }
  });

  const onSubmit = (data: LoginRequest) => {
    setErrorMessage(null);
    loginMutation.mutate(data);
  };

  return (
    <Container maxWidth="sm">
      <Paper elevation={3} sx={{ p: 4, mt: 8 }}>
        <Typography component="h1" variant="h5" align="center" gutterBottom>
          ログイン
        </Typography>
        {errorMessage && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {errorMessage}
          </Alert>
        )}
        <Box component="form" onSubmit={handleSubmit(onSubmit)} noValidate>
          <Controller
            name="email"
            control={control}
            rules={{
              required: 'メールアドレスを入力してください',
              pattern: {
                value: /^\S+@\S+\.\S+$/,
                message: '有効なメールアドレスを入力してください'
              }
            }}
            render={({ field }) => (
              <TextField
                margin="normal"
                required
                fullWidth
                id="email"
                label="メールアドレス"
                autoComplete="email"
                autoFocus
                error={!!errors.email}
                helperText={errors.email?.message}
                {...field}
              />
            )}
          />
          <Controller
            name="password"
            control={control}
            rules={{
              required: 'パスワードを入力してください',
              minLength: {
                value: 6,
                message: 'パスワードは6文字以上で入力してください'
              }
            }}
            render={({ field }) => (
              <TextField
                margin="normal"
                required
                fullWidth
                label="パスワード"
                type="password"
                id="password"
                autoComplete="current-password"
                error={!!errors.password}
                helperText={errors.password?.message}
                {...field}
              />
            )}
          />
          <Button
            type="submit"
            fullWidth
            variant="contained"
            sx={{ mt: 3, mb: 2 }}
            disabled={loginMutation.isLoading}
          >
            {loginMutation.isLoading ? (
              <CircularProgress size={24} color="inherit" />
            ) : (
              'ログイン'
            )}
          </Button>
          <Box sx={{ mt: 2, textAlign: 'center' }}>
            <Link component={RouterLink} to="/register" variant="body2">
              アカウントをお持ちでない方はこちら
            </Link>
          </Box>
        </Box>
      </Paper>
    </Container>
  );
};

export default LoginPage; 