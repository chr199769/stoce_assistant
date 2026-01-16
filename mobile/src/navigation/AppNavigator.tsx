import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import HomeScreen from '../screens/HomeScreen';
import PredictionScreen from '../screens/PredictionScreen';
import MarketAnalysisScreen from '../screens/MarketAnalysisScreen';
import SummaryScreen from '../screens/SummaryScreen';
import DragonTigerScreen from '../screens/DragonTigerScreen';
import SectorDetailScreen from '../screens/SectorDetailScreen';
import { Text } from 'react-native-paper';

// Define types for navigation
export type RootStackParamList = {
  Main: undefined;
  Prediction: { code?: string };
  SectorDetail: { sectorCode: string; sectorName: string };
  DragonTiger: undefined;
};

export type TabParamList = {
  Home: undefined;
  Summary: undefined;
  MarketAnalysis: undefined;
  DragonTigerTab: undefined;
};

const Tab = createBottomTabNavigator<TabParamList>();
const Stack = createNativeStackNavigator<RootStackParamList>();

const TabNavigator = () => {
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        headerShown: false,
        tabBarIcon: ({ color, size }) => {
          let iconName = '';
          if (route.name === 'Home') iconName = '🏠';
          else if (route.name === 'Summary') iconName = '📝';
          else if (route.name === 'MarketAnalysis') iconName = '🔮';
          else if (route.name === 'DragonTigerTab') iconName = '🐉';

          return <Text style={{ fontSize: size - 4, color }}>{iconName}</Text>;
        },
        tabBarActiveTintColor: '#1E88E5',
        tabBarInactiveTintColor: 'gray',
      })}
    >
      <Tab.Screen
        name="Home"
        component={HomeScreen}
        options={{ title: '首页' }}
      />
      <Tab.Screen
        name="Summary"
        component={SummaryScreen}
        options={{ title: '复盘' }}
      />
      <Tab.Screen
        name="MarketAnalysis"
        component={MarketAnalysisScreen}
        options={{ title: '盘前' }}
      />
      <Tab.Screen
        name="DragonTigerTab"
        component={DragonTigerScreen}
        options={{ title: '龙虎榜' }}
      />
    </Tab.Navigator>
  );
};

const AppNavigator = () => {
  return (
    <Stack.Navigator screenOptions={{ headerShown: false }}>
      <Stack.Screen name="Main" component={TabNavigator} />
      <Stack.Screen name="Prediction" component={PredictionScreen} />
      <Stack.Screen name="SectorDetail" component={SectorDetailScreen} />
      <Stack.Screen name="DragonTiger" component={DragonTigerScreen} />
    </Stack.Navigator>
  );
};

export default AppNavigator;
