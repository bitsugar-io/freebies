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
import { AppConfigProvider, useAppConfig } from './src/context/AppConfigContext';

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
  const { undismissedDeals, dealsLoading, userLoading } = useAppData();
  const { config } = useAppConfig();

  // Wait for data before choosing initial tab
  if (userLoading || dealsLoading) {
    return (
      <View style={[styles.loadingContainer, { backgroundColor: colors.background }]}>
        <ActivityIndicator size="large" color={colors.accent} />
      </View>
    );
  }

  // Maintenance mode — block the entire app
  if (config.features.maintenance_mode === true) {
    return (
      <View style={[styles.loadingContainer, { backgroundColor: colors.background }]}>
        <Text style={{ fontSize: 48, marginBottom: 16 }}>🔧</Text>
        <Text style={[styles.errorText, { color: colors.text }]}>
          We'll be right back
        </Text>
        <Text style={[styles.errorHint, { color: colors.textMuted }]}>
          Freebies is temporarily down for maintenance.{'\n'}Check back shortly!
        </Text>
      </View>
    );
  }

  return (
    <>
      <StatusBar style={isDark ? 'light' : 'dark'} />
      <Tab.Navigator
        initialRouteName={undismissedDeals.length > 0 ? 'Deals' : 'Discover'}
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
        <AppConfigProvider>
          <UserProvider>
            <AppDataProvider>
              <NavigationContainer>
                <MainApp />
              </NavigationContainer>
            </AppDataProvider>
          </UserProvider>
        </AppConfigProvider>
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
