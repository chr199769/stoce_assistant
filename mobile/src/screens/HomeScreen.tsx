import React, { useEffect, useState } from 'react';
import { View, StyleSheet, ScrollView, RefreshControl, Alert } from 'react-native';
import { Appbar, Card, Text, FAB, Dialog, Portal, TextInput, Button } from 'react-native-paper';
import { getRealtime } from '../api/stock';
import { RealtimeResponse } from '../types';
import { useNavigation } from '@react-navigation/native';
import AsyncStorage from '@react-native-async-storage/async-storage';

const WATCHLIST_KEY = 'stock_watchlist';

const HomeScreen = () => {
  const [stocks, setStocks] = useState<RealtimeResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [visible, setVisible] = useState(false);
  const [newCode, setNewCode] = useState('');
  const [watchlist, setWatchlist] = useState<string[]>([]);

  const navigation = useNavigation();

  // Load watchlist on mount
  useEffect(() => {
    loadWatchlist();
  }, []);

  const loadWatchlist = async () => {
    try {
      const stored = await AsyncStorage.getItem(WATCHLIST_KEY);
      if (stored) {
        setWatchlist(JSON.parse(stored));
      } else {
        // Default stocks
        const defaults = ['sh600519', 'sz000001', 'sh000001'];
        setWatchlist(defaults);
        await AsyncStorage.setItem(WATCHLIST_KEY, JSON.stringify(defaults));
      }
    } catch (e) {
      console.error('Failed to load watchlist', e);
    }
  };

  const saveWatchlist = async (newList: string[]) => {
    try {
      await AsyncStorage.setItem(WATCHLIST_KEY, JSON.stringify(newList));
      setWatchlist(newList);
    } catch (e) {
      console.error('Failed to save watchlist', e);
    }
  };

  const fetchStocks = async () => {
    if (watchlist.length === 0) return;
    setLoading(true);
    try {
      const promises = watchlist.map(code => getRealtime(code));
      const results = await Promise.all(promises);
      setStocks(results);
    } catch (error) {
      console.error(error);
      // Alert.alert('Error', 'Failed to fetch stock data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStocks();
  }, [watchlist]);

  const showDialog = () => setVisible(true);
  const hideDialog = () => setVisible(false);

  const addStock = () => {
    if (newCode) {
      if (!watchlist.includes(newCode)) {
        const newList = [...watchlist, newCode];
        saveWatchlist(newList);
      }
      setNewCode('');
      hideDialog();
    }
  };

  const getColor = (change: number) => {
    if (change > 0) return '#F44336'; // Red for up
    if (change < 0) return '#4CAF50'; // Green for down
    return '#333333';
  };

  return (
    <View style={styles.container}>
      <Appbar.Header style={styles.header}>
        <Appbar.Content title="股票助手" titleStyle={styles.headerTitle} />
      </Appbar.Header>

      <ScrollView
        contentContainerStyle={styles.scrollContent}
        refreshControl={<RefreshControl refreshing={loading} onRefresh={fetchStocks} />}
      >
        {stocks.map((stock) => (
          <Card key={stock.code} style={styles.card} onPress={() => {
             // Navigate to prediction with this code
             // @ts-ignore
             navigation.navigate('Prediction', { code: stock.code });
          }}>
            <Card.Content>
              <View style={styles.row}>
                <View>
                  <Text variant="titleMedium" style={styles.stockName}>{stock.name}</Text>
                  <Text variant="bodySmall" style={styles.stockCode}>{stock.code}</Text>
                </View>
                <View style={styles.priceContainer}>
                  <Text variant="titleLarge" style={{ color: getColor(stock.change_percent), fontWeight: 'bold' }}>
                    {stock.current_price.toFixed(2)}
                  </Text>
                  <Text variant="bodyMedium" style={{ color: getColor(stock.change_percent) }}>
                    {stock.change_percent > 0 ? '+' : ''}{stock.change_percent.toFixed(2)}%
                  </Text>
                </View>
              </View>
            </Card.Content>
          </Card>
        ))}
      </ScrollView>

      <Portal>
        <Dialog visible={visible} onDismiss={hideDialog}>
          <Dialog.Title>添加股票</Dialog.Title>
          <Dialog.Content>
            <TextInput
              label="股票代码 (如 sh600519)"
              value={newCode}
              onChangeText={setNewCode}
              mode="outlined"
            />
          </Dialog.Content>
          <Dialog.Actions>
            <Button onPress={hideDialog}>取消</Button>
            <Button onPress={addStock}>添加</Button>
          </Dialog.Actions>
        </Dialog>
      </Portal>

      <FAB
        icon="plus"
        style={styles.fab}
        onPress={showDialog}
        label="添加"
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  header: {
    backgroundColor: '#1E88E5',
  },
  headerTitle: {
    color: '#FFFFFF',
    fontWeight: 'bold',
  },
  scrollContent: {
    padding: 16,
    paddingBottom: 80,
  },
  card: {
    marginBottom: 12,
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    elevation: 2,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  stockName: {
    fontWeight: 'bold',
    fontSize: 16,
  },
  stockCode: {
    color: '#757575',
  },
  priceContainer: {
    alignItems: 'flex-end',
  },
  fab: {
    position: 'absolute',
    margin: 16,
    right: 0,
    bottom: 0,
    backgroundColor: '#1E88E5',
  },
});

export default HomeScreen;
