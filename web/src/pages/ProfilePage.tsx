import { Add as AddIcon, Delete as DeleteIcon, Edit as EditIcon } from '@mui/icons-material';
import {
    Alert,
    Box,
    Button,
    Card,
    CardContent,
    CircularProgress,
    Container,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Divider,
    Grid,
    IconButton,
    Paper,
    Tab,
    Tabs,
    TextField,
    Typography
} from '@mui/material';
import React, { useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { AddressApi, AddressRequest, AuthApi } from '../services/api';
import { useAuthStore } from '../store/authStore';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`profile-tabpanel-${index}`}
      aria-labelledby={`profile-tab-${index}`}
      {...other}
    >
      {value === index && (
        <Box sx={{ py: 3 }}>
          {children}
        </Box>
      )}
    </div>
  );
}

const ProfilePage: React.FC = () => {
  const { user, updateUser } = useAuthStore();
  const [tabValue, setTabValue] = useState(0);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [isAddressDialogOpen, setIsAddressDialogOpen] = useState(false);
  const [currentAddressId, setCurrentAddressId] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  // プロフィール編集フォーム
  const { control: profileControl, handleSubmit: handleProfileSubmit, formState: { errors: profileErrors } } = useForm({
    defaultValues: {
      firstName: user?.firstName || '',
      lastName: user?.lastName || '',
      phoneNumber: user?.phoneNumber || ''
    }
  });

  // 住所フォーム
  const { control: addressControl, handleSubmit: handleAddressSubmit, reset: resetAddressForm, formState: { errors: addressErrors } } = useForm<AddressRequest>({
    defaultValues: {
      streetAddress: '',
      city: '',
      state: '',
      postalCode: '',
      country: '日本',
      isDefault: false
    }
  });

  // プロフィール更新ミューテーション
  const updateProfileMutation = useMutation(AuthApi.updateProfile, {
    onSuccess: (data) => {
      updateUser(data);
      setSuccessMessage('プロフィールが更新されました');
      setTimeout(() => setSuccessMessage(null), 3000);
    },
    onError: (error: any) => {
      console.error('Profile update error:', error);
      setErrorMessage(
        error.response?.data?.message || 
        'プロフィールの更新に失敗しました。'
      );
    }
  });

  // 住所一覧取得クエリ
  const { data: addresses, isLoading: isLoadingAddresses } = useQuery(
    'addresses', 
    AddressApi.getAddresses,
    {
      onError: (error: any) => {
        console.error('Failed to fetch addresses:', error);
      }
    }
  );

  // 住所追加ミューテーション
  const addAddressMutation = useMutation(AddressApi.addAddress, {
    onSuccess: () => {
      queryClient.invalidateQueries('addresses');
      setIsAddressDialogOpen(false);
      resetAddressForm();
      setSuccessMessage('住所が追加されました');
      setTimeout(() => setSuccessMessage(null), 3000);
    },
    onError: (error: any) => {
      console.error('Address add error:', error);
      setErrorMessage(
        error.response?.data?.message || 
        '住所の追加に失敗しました。'
      );
    }
  });

  // 住所更新ミューテーション
  const updateAddressMutation = useMutation(
    ({ id, data }: { id: string, data: AddressRequest }) => 
      AddressApi.updateAddress(id, data),
    {
      onSuccess: () => {
        queryClient.invalidateQueries('addresses');
        setIsAddressDialogOpen(false);
        resetAddressForm();
        setCurrentAddressId(null);
        setSuccessMessage('住所が更新されました');
        setTimeout(() => setSuccessMessage(null), 3000);
      },
      onError: (error: any) => {
        console.error('Address update error:', error);
        setErrorMessage(
          error.response?.data?.message || 
          '住所の更新に失敗しました。'
        );
      }
    }
  );

  // 住所削除ミューテーション
  const deleteAddressMutation = useMutation(AddressApi.deleteAddress, {
    onSuccess: () => {
      queryClient.invalidateQueries('addresses');
      setSuccessMessage('住所が削除されました');
      setTimeout(() => setSuccessMessage(null), 3000);
    },
    onError: (error: any) => {
      console.error('Address delete error:', error);
      setErrorMessage(
        error.response?.data?.message || 
        '住所の削除に失敗しました。'
      );
    }
  });

  const onProfileSubmit = (data: any) => {
    setErrorMessage(null);
    updateProfileMutation.mutate(data);
  };

  const onAddressSubmit = (data: AddressRequest) => {
    setErrorMessage(null);
    if (currentAddressId) {
      updateAddressMutation.mutate({ id: currentAddressId, data });
    } else {
      addAddressMutation.mutate(data);
    }
  };

  const handleAddAddress = () => {
    resetAddressForm();
    setCurrentAddressId(null);
    setIsAddressDialogOpen(true);
  };

  const handleEditAddress = (address: any) => {
    resetAddressForm({
      streetAddress: address.streetAddress,
      city: address.city,
      state: address.state,
      postalCode: address.postalCode,
      country: address.country,
      isDefault: address.isDefault
    });
    setCurrentAddressId(address.id);
    setIsAddressDialogOpen(true);
  };

  const handleDeleteAddress = (addressId: string) => {
    if (window.confirm('この住所を削除してもよろしいですか？')) {
      deleteAddressMutation.mutate(addressId);
    }
  };

  return (
    <Container maxWidth="md">
      <Paper elevation={3} sx={{ p: 4, mt: 4 }}>
        <Typography component="h1" variant="h5" gutterBottom>
          マイページ
        </Typography>
        <Divider sx={{ mb: 2 }} />

        {errorMessage && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {errorMessage}
          </Alert>
        )}
        {successMessage && (
          <Alert severity="success" sx={{ mb: 2 }}>
            {successMessage}
          </Alert>
        )}

        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Tabs value={tabValue} onChange={handleTabChange} aria-label="プロフィールタブ">
            <Tab label="プロフィール情報" id="profile-tab-0" aria-controls="profile-tabpanel-0" />
            <Tab label="住所管理" id="profile-tab-1" aria-controls="profile-tabpanel-1" />
          </Tabs>
        </Box>

        <TabPanel value={tabValue} index={0}>
          <Box component="form" onSubmit={handleProfileSubmit(onProfileSubmit)} noValidate>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="firstName"
                  control={profileControl}
                  rules={{ required: '姓を入力してください' }}
                  render={({ field }) => (
                    <TextField
                      margin="normal"
                      required
                      fullWidth
                      id="firstName"
                      label="姓"
                      autoComplete="family-name"
                      error={!!profileErrors.firstName}
                      helperText={profileErrors.firstName?.message}
                      {...field}
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="lastName"
                  control={profileControl}
                  rules={{ required: '名を入力してください' }}
                  render={({ field }) => (
                    <TextField
                      margin="normal"
                      required
                      fullWidth
                      id="lastName"
                      label="名"
                      autoComplete="given-name"
                      error={!!profileErrors.lastName}
                      helperText={profileErrors.lastName?.message}
                      {...field}
                    />
                  )}
                />
              </Grid>
            </Grid>
            <Controller
              name="phoneNumber"
              control={profileControl}
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
                  label="電話番号"
                  autoComplete="tel"
                  error={!!profileErrors.phoneNumber}
                  helperText={profileErrors.phoneNumber?.message}
                  {...field}
                />
              )}
            />
            <Button
              type="submit"
              variant="contained"
              sx={{ mt: 3 }}
              disabled={updateProfileMutation.isLoading}
            >
              {updateProfileMutation.isLoading ? (
                <CircularProgress size={24} color="inherit" />
              ) : (
                '保存'
              )}
            </Button>
          </Box>
        </TabPanel>

        <TabPanel value={tabValue} index={1}>
          <Box sx={{ mb: 2, display: 'flex', justifyContent: 'flex-end' }}>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={handleAddAddress}
            >
              住所を追加
            </Button>
          </Box>
          {isLoadingAddresses ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', my: 4 }}>
              <CircularProgress />
            </Box>
          ) : addresses && addresses.length > 0 ? (
            <Grid container spacing={2}>
              {addresses.map((address: any) => (
                <Grid item xs={12} key={address.id}>
                  <Card>
                    <CardContent>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                        <Typography variant="h6" component="div">
                          {address.isDefault && (
                            <Typography component="span" variant="caption" sx={{ backgroundColor: 'primary.main', color: 'white', px: 1, py: 0.5, borderRadius: 1, mr: 1 }}>
                              デフォルト
                            </Typography>
                          )}
                          {address.streetAddress}
                        </Typography>
                        <Box>
                          <IconButton onClick={() => handleEditAddress(address)} size="small">
                            <EditIcon />
                          </IconButton>
                          <IconButton onClick={() => handleDeleteAddress(address.id)} size="small" color="error">
                            <DeleteIcon />
                          </IconButton>
                        </Box>
                      </Box>
                      <Typography variant="body1" color="text.secondary">
                        〒{address.postalCode}
                      </Typography>
                      <Typography variant="body1" color="text.secondary">
                        {address.state}{address.city}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        {address.country}
                      </Typography>
                    </CardContent>
                  </Card>
                </Grid>
              ))}
            </Grid>
          ) : (
            <Box sx={{ textAlign: 'center', py: 4 }}>
              <Typography variant="body1" color="text.secondary">
                登録された住所がありません
              </Typography>
              <Button variant="outlined" onClick={handleAddAddress} sx={{ mt: 2 }}>
                住所を追加する
              </Button>
            </Box>
          )}
        </TabPanel>
      </Paper>

      {/* 住所追加/編集ダイアログ */}
      <Dialog open={isAddressDialogOpen} onClose={() => setIsAddressDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {currentAddressId ? '住所を編集' : '住所を追加'}
        </DialogTitle>
        <DialogContent>
          <Box component="form" noValidate sx={{ mt: 1 }}>
            <Controller
              name="streetAddress"
              control={addressControl}
              rules={{ required: '住所を入力してください' }}
              render={({ field }) => (
                <TextField
                  margin="normal"
                  required
                  fullWidth
                  id="streetAddress"
                  label="住所 (番地など)"
                  autoComplete="street-address"
                  error={!!addressErrors.streetAddress}
                  helperText={addressErrors.streetAddress?.message}
                  {...field}
                />
              )}
            />
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="state"
                  control={addressControl}
                  rules={{ required: '都道府県を入力してください' }}
                  render={({ field }) => (
                    <TextField
                      margin="normal"
                      required
                      fullWidth
                      id="state"
                      label="都道府県"
                      error={!!addressErrors.state}
                      helperText={addressErrors.state?.message}
                      {...field}
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="city"
                  control={addressControl}
                  rules={{ required: '市区町村を入力してください' }}
                  render={({ field }) => (
                    <TextField
                      margin="normal"
                      required
                      fullWidth
                      id="city"
                      label="市区町村"
                      error={!!addressErrors.city}
                      helperText={addressErrors.city?.message}
                      {...field}
                    />
                  )}
                />
              </Grid>
            </Grid>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="postalCode"
                  control={addressControl}
                  rules={{ 
                    required: '郵便番号を入力してください',
                    pattern: {
                      value: /^[0-9]{3}-?[0-9]{4}$/,
                      message: '有効な郵便番号を入力してください'
                    }
                  }}
                  render={({ field }) => (
                    <TextField
                      margin="normal"
                      required
                      fullWidth
                      id="postalCode"
                      label="郵便番号"
                      placeholder="例: 123-4567"
                      error={!!addressErrors.postalCode}
                      helperText={addressErrors.postalCode?.message}
                      {...field}
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="country"
                  control={addressControl}
                  rules={{ required: '国を入力してください' }}
                  render={({ field }) => (
                    <TextField
                      margin="normal"
                      required
                      fullWidth
                      id="country"
                      label="国"
                      error={!!addressErrors.country}
                      helperText={addressErrors.country?.message}
                      {...field}
                    />
                  )}
                />
              </Grid>
            </Grid>
            <Controller
              name="isDefault"
              control={addressControl}
              render={({ field }) => (
                <Box sx={{ mt: 2 }}>
                  <Button
                    variant={field.value ? "contained" : "outlined"}
                    onClick={() => field.onChange(!field.value)}
                    size="small"
                  >
                    {field.value ? 'デフォルト住所に設定中' : 'デフォルト住所に設定する'}
                  </Button>
                </Box>
              )}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setIsAddressDialogOpen(false)}>
            キャンセル
          </Button>
          <Button
            onClick={handleAddressSubmit(onAddressSubmit)}
            variant="contained"
            disabled={addAddressMutation.isLoading || updateAddressMutation.isLoading}
          >
            {(addAddressMutation.isLoading || updateAddressMutation.isLoading) ? (
              <CircularProgress size={24} color="inherit" />
            ) : (
              '保存'
            )}
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default ProfilePage; 