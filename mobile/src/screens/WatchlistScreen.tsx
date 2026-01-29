import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, Alert } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import StockList from '../components/StockList';
import { getWatchlist } from '../api/stock';

interface WatchlistItem {
  stock_code: string;
  tags: string[];
  added_at: string;
}

const WatchlistScreen = () => {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<any[]>([]);

  useEffect(() => {
    fetchWatchlist();
  }, []);

  const fetchWatchlist = async () => {
    setLoading(true);
    try {
      const userId = await AsyncStorage.getItem('user_id');
      if (!userId) {
        // Alert.alert('Error', 'User not logged in');
        // Fallback for demo
        return;
      }

      const json = await getWatchlist(userId);
      if (json.items) {
        // 2. Ideally we need to fetch price info for each stock.
        // For now, we mock the price info or use the code as name.
        // In a real app, we would call a batch stock info API.
        const items = json.items.map((item: WatchlistItem) => ({
           symbol: item.stock_code,
           name: item.stock_code, // Placeholder
           price: 0,
           change: 0
        }));
        setData(items);
      }
    } catch (error) {
      console.error(error);
      // Alert.alert('Error', 'Failed to fetch watchlist');
    } finally {
      setLoading(false);
    }
  };

  const handlePress = (stock: any) => {
    // Navigate to detail?
    Alert.alert('Info', `Selected ${stock.symbol}`);
  };

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {data.length === 0 ? (
        <View style={styles.center}>
          <Text>No stocks in watchlist. Add some from Home screen.</Text>
        </View>
      ) : (
        <StockList data={data} onPress={handlePress} />
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#fff' },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center' },
});

export default WatchlistScreen;
