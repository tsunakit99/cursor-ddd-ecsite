import axios from 'axios';
import { useAuthStore } from '../store/authStore';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:50051';

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// リクエストインターセプター
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// レスポンスインターセプター
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    // 認証エラーの場合、ログアウト処理
    if (error.response && error.response.status === 401) {
      useAuthStore.getState().logout();
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  phoneNumber?: string;
}

export interface AddressRequest {
  streetAddress: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
  isDefault: boolean;
}

export interface LoginResponse {
  user: {
    id: string;
    email: string;
    firstName: string;
    lastName: string;
    phoneNumber?: string;
  };
  token: string;
}

export const AuthApi = {
  login: async (data: LoginRequest) => {
    const response = await api.post<LoginResponse>('/auth/login', data);
    return response.data;
  },
  register: async (data: RegisterRequest) => {
    const response = await api.post<LoginResponse>('/auth/register', data);
    return response.data;
  },
  getProfile: async () => {
    const response = await api.get('/users/profile');
    return response.data;
  },
  updateProfile: async (data: any) => {
    const response = await api.put('/users/profile', data);
    return response.data;
  },
};

export const AddressApi = {
  getAddresses: async () => {
    const response = await api.get('/users/addresses');
    return response.data;
  },
  addAddress: async (data: AddressRequest) => {
    const response = await api.post('/users/addresses', data);
    return response.data;
  },
  updateAddress: async (addressId: string, data: AddressRequest) => {
    const response = await api.put(`/users/addresses/${addressId}`, data);
    return response.data;
  },
  deleteAddress: async (addressId: string) => {
    const response = await api.delete(`/users/addresses/${addressId}`);
    return response.data;
  },
};

export default api; 