import React, { useEffect, useState, useCallback } from 'react';
import { View, StyleSheet, ScrollView, Alert, SafeAreaView, Platform, StatusBar } from 'react-native';
import { Card, Text, FAB, Dialog, Portal, TextInput, Button, IconButton } from 'react-native-paper';
import { getRealtime, recognizeStockImage } from '../api/stock';
import { RealtimeResponse } from '../types';
import { useNavigation, useFocusEffect } from '@react-navigation/native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { launchImageLibrary } from 'react-native-image-picker';

const WATCHLIST_KEY = 'stock_watchlist';

const HomeScreen = () => {
  const [stocks, setStocks] = useState<RealtimeResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [visible, setVisible] = useState(false);
  const [newCode, setNewCode] = useState('');
  const [watchlist, setWatchlist] = useState<string[]>([]);
  const [fabOpen, setFabOpen] = useState(false);

  const navigation = useNavigation();

  // Load watchlist on mount
  useEffect(() => {
    loadWatchlist();
  }, []);

  const loadWatchlist = async () => {
    try {
      const stored = await AsyncStorage.getItem(WATCHLIST_KEY);
      if (stored) {
        // De-duplicate on load
        const list = JSON.parse(stored);
        const uniqueList = Array.from(new Set(list)) as string[];
        setWatchlist(uniqueList);
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
      // Use Promise.allSettled to avoid entire failure if one stock fails
      const results = await Promise.allSettled(promises);

      const validStocks: RealtimeResponse[] = [];
      const seenCodes = new Set<string>();

      results.forEach((result, index) => {
        if (result.status === 'fulfilled') {
          const stock = result.value;
          // De-duplicate stocks based on code
          if (!seenCodes.has(stock.code)) {
            seenCodes.add(stock.code);
            validStocks.push(stock);
          }
        } else {
          console.error(`Failed to fetch stock ${watchlist[index]}:`, result.reason);
        }
      });

      setStocks(validStocks);
    } catch (error: any) {
      console.error('fetchStocks failed', error);
      // Alert.alert('Error', 'Failed to fetch stock data');
    } finally {
      setLoading(false);
    }
  };

  useFocusEffect(
    useCallback(() => {
      fetchStocks();
      const interval = setInterval(() => {
        fetchStocks(true);
      }, 3000);
      return () => clearInterval(interval);
    }, [watchlist])
  );

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

  const removeStock = (code: string) => {
    Alert.alert(
      '删除股票',
      '确定要删除这只股票吗？',
      [
        { text: '取消', style: 'cancel' },
        {
          text: '删除',
          style: 'destructive',
          onPress: () => {
            const newList = watchlist.filter(item => item !== code);
            saveWatchlist(newList);
          },
        },
      ]
    );
  };

  const handleImageImport = async () => {
    const result = await launchImageLibrary({
      mediaType: 'photo',
      selectionLimit: 1,
    });

    if (result.didCancel) return;
    if (result.errorCode) {
      Alert.alert('Error', result.errorMessage);
      return;
    }

    if (result.assets && result.assets.length > 0) {
      const asset = result.assets[0];
      setLoading(true);
      try {
        const response = await recognizeStockImage(asset.uri!, asset.type!, asset.fileName!);
        if (response.stocks.length > 0) {
          const newCodes = response.stocks.map(s => s.code).filter(c => !watchlist.includes(c));
          if (newCodes.length > 0) {
            const newList = [...watchlist, ...newCodes];
            saveWatchlist(newList);
            Alert.alert('成功', `已添加 ${newCodes.length} 只股票: ${response.stocks.map(s => `${s.name}(${s.code})`).join(', ')}`);
          } else {
            Alert.alert('提示', '未发现新股票或股票已在列表中');
          }
        } else {
          Alert.alert('提示', '未识别到股票信息');
        }
      } catch (e) {
        Alert.alert('错误', '图片识别失败');
        console.error(e);
      } finally {
        setLoading(false);
      }
    }
  };

  const getColor = (change: number) => {
    if (change > 0) return '#F44336'; // Red for up
    if (change < 0) return '#4CAF50'; // Green for down
    return '#333333';
  };

  return (
    <View style={styles.container}>
      <SafeAreaView style={styles.headerContainer}>
        <View style={styles.headerContent}>
          <Text style={styles.headerTitle}>股票助手</Text>
        </View>
      </SafeAreaView>

      <ScrollView
        contentContainerStyle={styles.scrollContent}
      >
        {stocks.map((stock) => (
          <Card key={stock.code} style={styles.card} onPress={() => {
            // Navigate to prediction with this code
            // @ts-ignore
            navigation.navigate('Prediction', { code: stock.code });
          }}>
            <Card.Content style={styles.cardContent}>
              <View style={styles.row}>
                <View>
                  <Text variant="titleMedium" style={styles.stockName}>{stock.name}</Text>
                  <Text variant="bodySmall" style={styles.stockCode}>{stock.code}</Text>
                </View>
                <View style={styles.rightGroup}>
                  <View style={styles.priceContainer}>
                    <Text variant="titleMedium" style={[styles.priceText, { color: getColor(stock.change_percent) }]}>
                      {stock.current_price.toFixed(2)}
                    </Text>
                    <Text variant="bodyMedium" style={[styles.percentText, { color: getColor(stock.change_percent) }]}>
                      {stock.change_percent > 0 ? '+' : ''}{stock.change_percent.toFixed(2)}%
                    </Text>
                  </View>
                  <Button
                    icon="delete-outline"
                    mode="text"
                    compact
                    textColor="#757575"
                    onPress={() => removeStock(stock.code)}
                  >
                    删除
                  </Button>
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

      <FAB.Group
        open={fabOpen}
        visible
        icon={fabOpen ? 'close' : 'plus'}
        actions={[
          { icon: 'plus', label: '手动添加', onPress: showDialog },
          { icon: 'image', label: '图片导入', onPress: handleImageImport },
        ]}
        onStateChange={({ open }) => setFabOpen(open)}
        style={styles.fab}
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  headerContainer: {
    backgroundColor: '#1E88E5',
    paddingTop: Platform.OS === 'android' ? StatusBar.currentHeight : 0,
  },
  headerContent: {
    height: 56,
    justifyContent: 'center',
    paddingHorizontal: 16,
    backgroundColor: '#1E88E5',
  },
  headerTitle: {
    color: '#FFFFFF',
    fontWeight: 'bold',
    fontSize: 20,
  },
  scrollContent: {
    padding: 16,
    paddingBottom: 80,
  },
  card: {
    marginBottom: 8,
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    elevation: 2,
  },
  cardContent: {
    paddingVertical: 8,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  rightGroup: {
    flexDirection: 'row',
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
    flexDirection: 'row',
    alignItems: 'center',
  },
  priceText: {
    fontWeight: 'bold',
  },
  percentText: {
    marginLeft: 8,
  },
  fab: {
    position: 'absolute',
    margin: 16,
    right: 0,
    bottom: 0,
  },
});

export default HomeScreen;
