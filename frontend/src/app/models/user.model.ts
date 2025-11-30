export interface User {
  id: string;
  email: string;
  name: string;
  contact?: string;
  preferences: UserPreferences;
}

export interface UserPreferences {
  currency: string;
  theme: 'light' | 'dark';
  notifications: boolean;
}

export interface AuthRequest {
  email: string;
  password: string;
}

export interface SignupRequest extends AuthRequest {
  name: string;
  confirmPassword: string;
}