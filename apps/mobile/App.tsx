import React from 'react';
import { StyleSheet, Text, View, ActivityIndicator } from 'react-native';
import { StatusBar } from 'expo-status-bar';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { SafeAreaProvider } from 'react-native-safe-area-context';

import { DealsScreen } from './src/screens/DealsScreen';
import { DiscoverScreen } from './src/screens/DiscoverScreen';
import { ProfileScreen } from './src/screens/ProfileScreen';
import { UserProvider, useUser } from './src/hooks/useUser';
import { ThemeProvider, useTheme } from './src/hooks/useTheme';
import { AppDataProvider, useAppData } from './src/context/AppDataContext';

const Tab = createBottomTabNavigator();

function TabIcon({ name, focused, color }: { name: string; focused: boolean; color: string }) {
  const icons: Record<string, string> = {
    Deals: '🎁',
    Discover: '🔍',
    Profile: '👤',
  };
  return (
    <Text style={{ fontSize: focused ? 26 : 24, opacity: focused ? 1 : 0.7 }}>
      {icons[name]}
    </Text>
  );
}

function MainApp() {
  const { theme } = useTheme();
  const { colors, isDark } = theme;
  const { userLoading, userError, eventsLoading, eventsError, undismissedDeals } = useAppData();

  const isLoading = userLoading || eventsLoading;
  const error = userError || eventsError;

  if (isLoading) {
    return (
      <View style={[styles.loadingContainer, { backgroundColor: colors.background }]}>
        <ActivityIndicator size="large" color={colors.accent} />
        <Text style={[styles.loadingText, { color: colors.textMuted }]}>
          Loading...
        </Text>
      </View>
    );
  }

  if (error) {
    return (
      <View style={[styles.loadingContainer, { backgroundColor: colors.background }]}>
        <Text style={[styles.errorText, { color: colors.warning }]}>
          {error}
        </Text>
        <Text style={[styles.errorHint, { color: colors.textMuted }]}>
          Make sure the API server is running:{'\n'}
          task api:serve
        </Text>
      </View>
    );
  }

  return (
    <>
      <StatusBar style={isDark ? 'light' : 'dark'} />
      <Tab.Navigator
        screenOptions={({ route }) => ({
          tabBarIcon: ({ focused, color }) => (
            <TabIcon name={route.name} focused={focused} color={color} />
          ),
          tabBarActiveTintColor: colors.accent,
          tabBarInactiveTintColor: colors.textMuted,
          tabBarStyle: {
            backgroundColor: colors.surface,
            borderTopColor: colors.border,
            paddingTop: 8,
            paddingBottom: 8,
            height: 60,
          },
          tabBarLabelStyle: {
            fontSize: 12,
            fontWeight: '500',
          },
          headerStyle: {
            backgroundColor: colors.surface,
          },
          headerTintColor: colors.text,
          headerTitleStyle: {
            fontWeight: '600',
          },
        })}
      >
        <Tab.Screen
          name="Deals"
          component={DealsScreen}
          options={{
            tabBarBadge: undismissedDeals.length > 0 ? undismissedDeals.length : undefined,
            tabBarBadgeStyle: { backgroundColor: colors.accent },
            headerTitle: 'Active Deals',
          }}
        />
        <Tab.Screen
          name="Discover"
          component={DiscoverScreen}
          options={{
            headerTitle: 'Discover Events',
          }}
        />
        <Tab.Screen
          name="Profile"
          component={ProfileScreen}
          options={{
            headerTitle: 'Profile & Settings',
          }}
        />
      </Tab.Navigator>
    </>
  );
}

export default function App() {
  return (
    <SafeAreaProvider>
      <ThemeProvider>
        <UserProvider>
          <AppDataProvider>
            <NavigationContainer>
              <MainApp />
            </NavigationContainer>
          </AppDataProvider>
        </UserProvider>
      </ThemeProvider>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  loadingContainer: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 20,
  },
  loadingText: {
    marginTop: 12,
    fontSize: 16,
  },
  errorText: {
    fontSize: 16,
    fontWeight: '600',
    textAlign: 'center',
    marginBottom: 12,
  },
  errorHint: {
    fontSize: 14,
    textAlign: 'center',
    fontFamily: 'monospace',
  },
});
