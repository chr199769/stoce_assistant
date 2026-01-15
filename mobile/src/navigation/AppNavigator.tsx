import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import HomeScreen from '../screens/HomeScreen';
import PredictionScreen from '../screens/PredictionScreen';
import SummaryScreen from '../screens/SummaryScreen';
import SectorDetailScreen from '../screens/SectorDetailScreen';
import DragonTigerScreen from '../screens/DragonTigerScreen';
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
  PredictionTab: undefined;
  Summary: undefined;
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
          if (route.name === 'Home') iconName = '🏠'; // Fallback to emoji if icons fail
          else if (route.name === 'PredictionTab') iconName = '🔮';
          else if (route.name === 'Summary') iconName = '📊';
          else if (route.name === 'DragonTigerTab') iconName = '🐉';

          // Using Text as icon to avoid linking issues for now.
          // In a real app, use <Icon source="home" color={color} size={size} /> from react-native-paper
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
        options={{ title: '大盘总结' }}
      />
      <Tab.Screen
        name="DragonTigerTab"
        component={DragonTigerScreen}
        options={{ title: '龙虎榜' }}
      />
      <Tab.Screen
        name="PredictionTab"
        component={PredictionScreen}
        options={{ title: '个股预测' }}
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
