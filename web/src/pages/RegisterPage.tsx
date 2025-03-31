import {
    Alert,
    Box,
    Button,
    CircularProgress,
    Container,
    Grid,
    Link,
    Paper,
    TextField,
    Typography
} from '@mui/material';
import React, { useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useMutation } from 'react-query';
import { Link as RouterLink, useNavigate } from 'react-router-dom';
import { AuthApi, RegisterRequest } from '../services/api';
import { useAuthStore } from '../store/authStore';

const RegisterPage: React.FC = () => {
  const navigate = useNavigate();
  const { login } = useAuthStore();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const { control, handleSubmit, formState: { errors } } = useForm<RegisterRequest>({
    defaultValues: {
      email: '',
      password: '',
      firstName: '',
      lastName: '',
      phoneNumber: ''
    }
  });

  const registerMutation = useMutation(AuthApi.register, {
    onSuccess: (data) => {
      login(data.user, data.token);
      navigate('/');
    },
    onError: (error: any) => {
      console.error('Registration error:', error);
      setErrorMessage(
        error.response?.data?.message || 
        'ユーザー登録に失敗しました。入力内容を確認してください。'
      );
    }
  });

  const onSubmit = (data: RegisterRequest) => {
    setErrorMessage(null);
    registerMutation.mutate(data);
  };

  return (
    <Container maxWidth="sm">
      <Paper elevation={3} sx={{ p: 4, mt: 8 }}>
        <Typography component="h1" variant="h5" align="center" gutterBottom>
          アカウント登録
        </Typography>
        {errorMessage && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {errorMessage}
          </Alert>
        )}
        <Box component="form" onSubmit={handleSubmit(onSubmit)} noValidate>
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6}>
              <Controller
                name="firstName"
                control={control}
                rules={{
                  required: '姓を入力してください'
                }}
                render={({ field }) => (
                  <TextField
                    margin="normal"
                    required
                    fullWidth
                    id="firstName"
                    label="姓"
                    autoComplete="family-name"
                    error={!!errors.firstName}
                    helperText={errors.firstName?.message}
                    {...field}
                  />
                )}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <Controller
                name="lastName"
                control={control}
                rules={{
                  required: '名を入力してください'
                }}
                render={({ field }) => (
                  <TextField
                    margin="normal"
                    required
                    fullWidth
                    id="lastName"
                    label="名"
                    autoComplete="given-name"
                    error={!!errors.lastName}
                    helperText={errors.lastName?.message}
                    {...field}
                  />
                )}
              />
            </Grid>
          </Grid>
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
                autoComplete="new-password"
                error={!!errors.password}
                helperText={errors.password?.message}
                {...field}
              />
            )}
          />
          <Controller
            name="phoneNumber"
            control={control}
            rules={{
              pattern: {
                value: /^[0-9-]+$/,
                message: '有効な電話番号を入力してください'
              }
            }}
            render={({ field }) => (
              <TextField
                margin="normal"
                fullWidth
                id="phoneNumber"
                label="電話番号 (任意)"
                autoComplete="tel"
                error={!!errors.phoneNumber}
                helperText={errors.phoneNumber?.message}
                {...field}
              />
            )}
          />
          <Button
            type="submit"
            fullWidth
            variant="contained"
            sx={{ mt: 3, mb: 2 }}
            disabled={registerMutation.isLoading}
          >
            {registerMutation.isLoading ? (
              <CircularProgress size={24} color="inherit" />
            ) : (
              '登録'
            )}
          </Button>
          <Box sx={{ mt: 2, textAlign: 'center' }}>
            <Link component={RouterLink} to="/login" variant="body2">
              すでにアカウントをお持ちの方はこちら
            </Link>
          </Box>
        </Box>
      </Paper>
    </Container>
  );
};

export default RegisterPage; 