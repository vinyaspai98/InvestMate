export interface NotificationPreferences {
  email: boolean;
  push: boolean;
}

export interface UserPreferences {
  currency: string;
  theme: 'light' | 'dark';
  notifications: boolean | NotificationPreferences;
}

export interface User {
  id: string;
  email: string;
  name: string;
  phoneNumber?: string;
  contact?: string; // compatibility alias for phoneNumber
  preferences: UserPreferences;
}