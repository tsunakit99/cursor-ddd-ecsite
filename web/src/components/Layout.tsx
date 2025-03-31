import {
    AppBar,
    Box,
    Button,
    Container,
    Link,
    Toolbar,
    Typography
} from '@mui/material';
import React from 'react';
import { Outlet, Link as RouterLink, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';

const Layout: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated, logout } = useAuthStore();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            <Link 
              component={RouterLink} 
              to="/" 
              color="inherit" 
              underline="none"
            >
              ECサイト
            </Link>
          </Typography>
          <Box sx={{ display: 'flex', gap: 2 }}>
            {isAuthenticated ? (
              <>
                <Button 
                  color="inherit" 
                  component={RouterLink} 
                  to="/profile"
                >
                  プロフィール
                </Button>
                <Button 
                  color="inherit" 
                  onClick={handleLogout}
                >
                  ログアウト
                </Button>
              </>
            ) : (
              <>
                <Button 
                  color="inherit" 
                  component={RouterLink} 
                  to="/login"
                >
                  ログイン
                </Button>
                <Button 
                  color="inherit" 
                  component={RouterLink} 
                  to="/register"
                >
                  登録
                </Button>
              </>
            )}
          </Box>
        </Toolbar>
      </AppBar>
      <Container component="main" sx={{ flex: '1 0 auto', py: 4 }}>
        <Outlet />
      </Container>
      <Box 
        component="footer" 
        sx={{ 
          py: 3, 
          px: 2, 
          mt: 'auto', 
          backgroundColor: (theme) => theme.palette.grey[200] 
        }}
      >
        <Container maxWidth="sm">
          <Typography variant="body2" color="text.secondary" align="center">
            © {new Date().getFullYear()} ECサイト All rights reserved.
          </Typography>
        </Container>
      </Box>
    </Box>
  );
};

export default Layout; 