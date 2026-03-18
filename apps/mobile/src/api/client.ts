// API client for Freebies backend
import * as SecureStore from 'expo-secure-store';

// Configure this based on environment
// const API_BASE_URL = 'https://freebie-api.bitsugar.io/api/v1';
const API_BASE_URL = 'http://localhost:8080/api/v1';

const AUTH_TOKEN_KEY = 'freebie_auth_token';

export interface League {
  id: string;
  name: string;
  icon: string;
  displayOrder: number;
}

export interface Event {
  id: string;
  offerId: string;
  teamId: string;
  teamName: string;
  league: string;
  teamColor?: string;
  icon?: string;
  partnerName: string;
  offerName: string;
  offerDescription: string;
  triggerCondition: string;
  regionCode?: string;
  regionName?: string;
  offerUrl?: string;
  affiliateUrl?: string;
  affiliateTagline?: string;
  isActive: boolean;
}

export interface User {
  id: string;
  deviceId: string;
  platform: string;
  pushToken?: string;
  token?: string; // Auth token, only returned on creation
}

export interface Subscription {
  id: string;
  userId: string;
  eventId: string;
  event: Event;
}

export interface ActiveDeal {
  id: string;
  eventId: string;
  triggeredAt: string;
  expiresAt?: string;
  event: Event;
  isDismissed: boolean;
  dismissalType?: 'got_it' | 'stop_reminding';
}

export interface Dismissal {
  id: string;
  userId: string;
  triggeredEventId: string;
  type: 'got_it' | 'stop_reminding';
  dismissedAt: string;
}

export interface ScreenBlock {
  type: string;
  key: string;
  config: Record<string, any>;
}

export interface AppConfig {
  features: Record<string, boolean>;
  screens: Record<string, ScreenBlock[]>;
}

export interface ApiError {
  error: string;
}

class ApiClient {
  private baseUrl: string;
  private authToken: string | null = null;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  // Load auth token from secure storage
  async loadAuthToken(): Promise<string | null> {
    this.authToken = await SecureStore.getItemAsync(AUTH_TOKEN_KEY);
    return this.authToken;
  }

  // Save auth token to secure storage
  async setAuthToken(token: string): Promise<void> {
    this.authToken = token;
    await SecureStore.setItemAsync(AUTH_TOKEN_KEY, token);
  }

  // Clear auth token
  async clearAuthToken(): Promise<void> {
    this.authToken = null;
    await SecureStore.deleteItemAsync(AUTH_TOKEN_KEY);
  }

  // Get current auth token
  getAuthToken(): string | null {
    return this.authToken;
  }

  private async request<T>(
    path: string,
    options: RequestInit = {},
    authenticated: boolean = false
  ): Promise<T> {
    const url = `${this.baseUrl}${path}`;
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    // Add auth header for authenticated requests
    if (authenticated && this.authToken) {
      headers['Authorization'] = `Bearer ${this.authToken}`;
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({
        error: `HTTP ${response.status}`,
      }));
      throw new Error(error.error || `Request failed: ${response.status}`);
    }

    // Handle 204 No Content
    if (response.status === 204) {
      return undefined as T;
    }

    return response.json();
  }

  // Leagues
  async listLeagues(): Promise<League[]> {
    return this.request<League[]>('/leagues');
  }

  // Events
  async listEvents(): Promise<Event[]> {
    return this.request<Event[]>('/events');
  }

  async getEvent(id: string): Promise<Event> {
    return this.request<Event>(`/events/${id}`);
  }

  // Users
  async createUser(deviceId: string, platform: string): Promise<User> {
    const user = await this.request<User>('/users', {
      method: 'POST',
      body: JSON.stringify({ deviceId, platform }),
    });
    // Store token for future authenticated requests
    if (user.token) {
      await this.setAuthToken(user.token);
    }
    return user;
  }

  async getUser(id: string): Promise<User> {
    return this.request<User>(`/users/${id}`, {}, true);
  }

  async updatePushToken(userId: string, pushToken: string): Promise<void> {
    return this.request<void>(
      `/users/${userId}/push-token`,
      {
        method: 'PUT',
        body: JSON.stringify({ pushToken }),
      },
      true
    );
  }

  // Subscriptions
  async listSubscriptions(userId: string): Promise<Subscription[]> {
    return this.request<Subscription[]>(`/users/${userId}/subscriptions`, {}, true);
  }

  async createSubscription(
    userId: string,
    eventId: string
  ): Promise<Subscription> {
    return this.request<Subscription>(
      `/users/${userId}/subscriptions`,
      {
        method: 'POST',
        body: JSON.stringify({ eventId }),
      },
      true
    );
  }

  async deleteSubscription(userId: string, eventId: string): Promise<void> {
    return this.request<void>(
      `/users/${userId}/subscriptions/${eventId}`,
      { method: 'DELETE' },
      true
    );
  }

  // Active Deals
  async listActiveDeals(userId: string): Promise<ActiveDeal[]> {
    return this.request<ActiveDeal[]>(`/users/${userId}/active-deals`, {}, true);
  }

  // Dismissals
  async createDismissal(
    userId: string,
    triggeredEventId: string,
    type: 'got_it' | 'stop_reminding' = 'got_it'
  ): Promise<Dismissal> {
    return this.request<Dismissal>(
      `/users/${userId}/dismissals`,
      {
        method: 'POST',
        body: JSON.stringify({ triggeredEventId, type }),
      },
      true
    );
  }

  async deleteDismissal(userId: string, triggeredEventId: string): Promise<void> {
    return this.request<void>(
      `/users/${userId}/dismissals/${triggeredEventId}`,
      { method: 'DELETE' },
      true
    );
  }

  // User Stats
  async getUserStats(userId: string): Promise<UserStats> {
    return this.request<UserStats>(`/users/${userId}/stats`, {}, true);
  }

  // App Config
  async getConfig(): Promise<AppConfig> {
    return this.request<AppConfig>('/config');
  }
}

export interface UserStats {
  dealsClaimed: number;
  subscriptionsCount: number;
}

export const api = new ApiClient();
