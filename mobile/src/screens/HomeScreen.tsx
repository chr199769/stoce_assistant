import React, { useEffect, useState, useCallback, useRef } from 'react';
import { View, StyleSheet, ScrollView, Alert, SafeAreaView, Platform, StatusBar, TouchableOpacity } from 'react-native';
import { Card, Text, FAB, Dialog, Portal, TextInput, Button, ActivityIndicator, Divider } from 'react-native-paper';
import { getRealtime, recognizeStockImage, getPrediction } from '../api/stock';
import { RealtimeResponse, PredictionResponse } from '../types';
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
  const watchlistRef = useRef<string[]>([]);
  const [fabOpen, setFabOpen] = useState(false);
  const [predictions, setPredictions] = useState<{[key: string]: PredictionResponse}>({});
  const [predicting, setPredicting] = useState<{[key: string]: boolean}>({});
  const [expandedStock, setExpandedStock] = useState<string | null>(null);

  const navigation = useNavigation();

  // Keep ref in sync with state
  useEffect(() => {
    watchlistRef.current = watchlist;
  }, [watchlist]);

  // Load watchlist on mount
  useEffect(() => {
    loadWatchlist();
  }, []);

  const loadWatchlist = async () => {
    try {
      const stored = await AsyncStorage.getItem(WATCHLIST_KEY);
      if (stored) {
        const list = JSON.parse(stored);
        // Clean up data: trim whitespace, convert to lowercase, and de-duplicate
        const cleanList = Array.from(new Set(
          list.map((item: string) => item.trim().toLowerCase())
        )) as string[];

        // Filter out empty strings
        const validList = cleanList.filter(item => item.length > 0);

        setWatchlist(validList);
        // Save cleaned list back to storage if it changed
        if (JSON.stringify(list) !== JSON.stringify(validList)) {
          saveWatchlist(validList);
        }
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
    if (watchlist.length === 0) {
      setStocks([]);
      return;
    }
    setLoading(true);
    try {
      const promises = watchlist.map(code => getRealtime(code));
      // Use Promise.allSettled to avoid entire failure if one stock fails
      const results = await Promise.allSettled(promises);

      const validStocks: RealtimeResponse[] = [];
      const seenCodes = new Set<string>();

      // Get the latest watchlist from ref to avoid race conditions with delete operations
      const currentWatchlistSet = new Set(watchlistRef.current.map(c => c.trim().toLowerCase()));

      results.forEach((result, index) => {
        if (result.status === 'fulfilled') {
          const stock = result.value;
          const stockCode = stock.code.trim().toLowerCase();

          // Only add if it's still in the watchlist
          if (currentWatchlistSet.has(stockCode)) {
            // De-duplicate stocks based on code
            if (!seenCodes.has(stockCode)) {
              seenCodes.add(stockCode);
              validStocks.push(stock);
            }
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

  const isMarketOpen = () => {
    const now = new Date();
    const day = now.getDay();
    const hour = now.getHours();
    const minute = now.getMinutes();

    // Weekend check (0 is Sunday, 6 is Saturday)
    if (day === 0 || day === 6) return false;

    // Time check: 09:15 - 11:30, 13:00 - 15:05
    const time = hour * 100 + minute;
    return (time >= 915 && time <= 1130) || (time >= 1300 && time <= 1505);
  };

  useFocusEffect(
    useCallback(() => {
      fetchStocks();
      const interval = setInterval(() => {
        if (isMarketOpen()) {
          fetchStocks();
        }
      }, 3000);
      return () => clearInterval(interval);
    }, [watchlist])
  );

  const showDialog = () => setVisible(true);
  const hideDialog = () => setVisible(false);

  const addStock = () => {
    if (newCode) {
      const code = newCode.trim().toLowerCase();
      if (code && !watchlist.includes(code)) {
        const newList = [...watchlist, code];
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
            // Trim and lowercase for robust comparison
            const targetCode = code.trim().toLowerCase();
            const newList = watchlist.filter(item => item.trim().toLowerCase() !== targetCode);
            saveWatchlist(newList);
            // Optimistically update stocks state to remove the item immediately
            setStocks(prev => prev.filter(s => s.code !== code));
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
          const newCodes = response.stocks.map(s => s.code.trim().toLowerCase()).filter(c => !watchlist.includes(c));
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
      } catch (e: any) {
        const errorMessage = e.response?.data || e.message || '图片识别失败';
        Alert.alert('错误', `图片识别失败: ${errorMessage}`);
        console.error(e);
      } finally {
        setLoading(false);
      }
    }
  };

  const handlePredict = async (code: string) => {
    if (expandedStock === code) {
      setExpandedStock(null);
      return;
    }
    
    setExpandedStock(code);
    
    if (predictions[code]) return; // Already have data

    setPredicting(prev => ({ ...prev, [code]: true }));
    try {
      const res = await getPrediction({
        code,
        days: 3,
        include_news: true,
        model: 'doubao-pro-32k'
      });
      setPredictions(prev => ({ ...prev, [code]: res }));
    } catch (error) {
      console.error('Prediction failed', error);
      Alert.alert('错误', '获取预测失败');
    } finally {
      setPredicting(prev => ({ ...prev, [code]: false }));
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
          <Card key={stock.code} style={styles.card}>
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

              <View style={styles.actionRow}>
                <Button 
                  mode={expandedStock === stock.code ? "contained-tonal" : "outlined"} 
                  onPress={() => handlePredict(stock.code)}
                  compact
                  icon="crystal-ball"
                  style={styles.predictBtn}
                >
                  AI 预测
                </Button>
              </View>

              {expandedStock === stock.code && (
                <View style={styles.predictionContainer}>
                  <Divider style={styles.divider} />
                  {predicting[stock.code] ? (
                    <ActivityIndicator animating={true} size="small" style={styles.loader} />
                  ) : predictions[stock.code] ? (
                    <View>
                      <View style={styles.predictionHeader}>
                        <Text variant="labelLarge">置信度: {(predictions[stock.code].confidence * 100).toFixed(0)}%</Text>
                      </View>
                      <Text variant="bodyMedium" style={styles.analysisText}>
                        {predictions[stock.code].analysis}
                      </Text>
                      {predictions[stock.code].news_summary && (
                         <View style={styles.newsBox}>
                           <Text variant="bodySmall" style={styles.newsText}>
                             📰 {predictions[stock.code].news_summary}
                           </Text>
                         </View>
                      )}
                    </View>
                  ) : (
                    <Text>暂无数据</Text>
                  )}
                </View>
              )}
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
  actionRow: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    marginTop: 8,
  },
  predictBtn: {
    marginLeft: 8,
  },
  predictionContainer: {
    marginTop: 8,
    paddingTop: 8,
  },
  divider: {
    marginBottom: 8,
  },
  loader: {
    marginVertical: 10,
  },
  predictionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 4,
  },
  analysisText: {
    lineHeight: 20,
    color: '#333',
  },
  newsBox: {
    marginTop: 8,
    padding: 8,
    backgroundColor: '#E3F2FD',
    borderRadius: 4,
  },
  newsText: {
    color: '#1565C0',
  },
});

export default HomeScreen;
