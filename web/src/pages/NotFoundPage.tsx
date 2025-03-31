import {
    Box,
    Button,
    Container,
    Paper,
    Typography
} from '@mui/material';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';

const NotFoundPage: React.FC = () => {
  return (
    <Container maxWidth="sm">
      <Paper elevation={3} sx={{ p: 4, mt: 8, textAlign: 'center' }}>
        <Typography variant="h1" component="h1" gutterBottom>
          404
        </Typography>
        <Typography variant="h5" component="h2" gutterBottom>
          ページが見つかりません
        </Typography>
        <Typography variant="body1" color="text.secondary" paragraph>
          お探しのページは存在しないか、移動した可能性があります。
        </Typography>
        <Box sx={{ mt: 4 }}>
          <Button 
            variant="contained" 
            component={RouterLink} 
            to="/"
            size="large"
          >
            ホームに戻る
          </Button>
        </Box>
      </Paper>
    </Container>
  );
};

export default NotFoundPage; 