import React, { useEffect, useState, useCallback, useRef } from 'react';
import { View, StyleSheet, ScrollView, Alert, SafeAreaView, Platform, StatusBar, TouchableOpacity } from 'react-native';
import { Card, Text, FAB, Dialog, Portal, TextInput, Button, ActivityIndicator, Divider, SegmentedButtons, Chip } from 'react-native-paper';
import { getRealtime, recognizeStockImage, getPrediction, addWatchlist, removeWatchlist, getWatchlist } from '../api/stock';
import { RealtimeResponse, PredictionResponse } from '../types';
import { useNavigation, useFocusEffect } from '@react-navigation/native';
import { useAuth } from '../context/AuthContext';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { launchImageLibrary } from 'react-native-image-picker';
import FractalChart, { FractalData } from '../components/FractalChart';

const WATCHLIST_KEY = 'stock_watchlist';

const HomeScreen = () => {
  const [stocks, setStocks] = useState<RealtimeResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [visible, setVisible] = useState(false);
  const [newCode, setNewCode] = useState('');
  const [watchlist, setWatchlist] = useState<string[]>([]);
  const watchlistRef = useRef<string[]>([]);
  const [fabOpen, setFabOpen] = useState(false);
  const [predictions, setPredictions] = useState<{ [key: string]: PredictionResponse }>({});
  const [predicting, setPredicting] = useState<{ [key: string]: boolean }>({});
  const [expandedStock, setExpandedStock] = useState<string | null>(null);
  const [singlePredictCode, setSinglePredictCode] = useState('');
  const [singlePredictLoading, setSinglePredictLoading] = useState(false);
  const [predictionDialogVisible, setPredictionDialogVisible] = useState(false);
  const [currentPrediction, setCurrentPrediction] = useState<PredictionResponse | null>(null);
  const [currentFractalData, setCurrentFractalData] = useState<FractalData | null>(null);
  const [selectedModel, setSelectedModel] = useState('glm-4.6v-flash');

  const { user } = useAuth();
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
      if (user) {
        console.log('Loading remote watchlist for user:', user.id);
        const res = await getWatchlist(user.id);
        console.log('Got remote watchlist:', res);

        // Ensure we are getting an array
        const items = res.items || [];
        const list = items.map(i => i.stock_code);
        setWatchlist(list);
      } else {
        setWatchlist([]);
      }
    } catch (e) {
      console.error('Failed to load watchlist', e);
      setWatchlist([]);
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
    // If user is logged in, merge local watchlist with remote
    let currentList = [...watchlist];

    // Always fetch, even if watchlist is empty (to clear list)
    if (currentList.length === 0) {
      setStocks([]);
      return;
    }

    setLoading(true);
    try {
      const promises = currentList.map(code => getRealtime(code));
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

  const addStock = async () => {
    console.log('addStock called with:', newCode);
    console.log('Current User:', user);

    if (newCode) {
      const code = newCode.trim().toLowerCase();
      if (!code) return;

      // 1. Try to sync with backend if logged in (regardless of local existence)
      if (user) {
        try {
          console.log('Sending AddWatchlist request to backend for:', code);
          await addWatchlist(user.id, code);
          console.log('Backend request success');

          // Refresh list from backend
          await loadWatchlist();
        } catch (error) {
          console.error('Failed to sync stock to backend:', error);
          Alert.alert('错误', '添加失败，请重试');
        }
      } else {
        console.warn('User is null, skipping backend sync');
        Alert.alert('提示', '请先登录');
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
          onPress: async () => {
            try {
              if (user) {
                await removeWatchlist(user.id, code);
                // Refresh list from backend
                await loadWatchlist();
              }
            } catch (e) {
              console.error('Failed to remove stock', e);
              Alert.alert('错误', '删除失败');
            }
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
            if (user) {
              // Add all to backend
              await Promise.all(newCodes.map(c => addWatchlist(user.id, c)));
            }
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

  const parseMetadata = (analysis: string) => {
    const parts = analysis.split('---METADATA---');
    if (parts.length > 1) {
      try {
        const metadataJson = parts[1].trim();
        const metadata = JSON.parse(metadataJson);
        return {
            text: parts[0].trim(),
            fractalData: metadata.fractal_data || null
        };
      } catch (e) {
        console.error("Failed to parse metadata", e);
      }
    }
    return { text: analysis, fractalData: null };
  };

  const handlePredict = async (code: string) => {
    // Navigate to Prediction Screen directly with the selected model
    // @ts-ignore
    navigation.navigate('Prediction', { code, model: selectedModel });
  };

  const handleSinglePredict = async () => {
    if (!singlePredictCode) return;
    const code = singlePredictCode.trim().toLowerCase();
    if (!code) return;
    
    // Use the same handlePredict logic but clear input
    await handlePredict(code);
    setSinglePredictCode('');
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
        <View style={styles.predictInputContainer}>
          <TextInput
            placeholder="输入股票代码预测 (如 sh600519)"
            value={singlePredictCode}
            onChangeText={setSinglePredictCode}
            mode="outlined"
            style={styles.predictInput}
            dense
            right={<TextInput.Icon icon="crystal-ball" onPress={handleSinglePredict} />}
            onSubmitEditing={handleSinglePredict}
          />
          <View style={styles.modelSelectorContainer}>
            <ScrollView horizontal showsHorizontalScrollIndicator={false}>
                <Chip 
                    selected={selectedModel === 'glm-4.6v-flash'} 
                    onPress={() => setSelectedModel('glm-4.6v-flash')}
                    style={styles.modelChip}
                    icon="brain"
                >深度分析</Chip>
                <Chip 
                    selected={selectedModel === 'fractal'} 
                    onPress={() => setSelectedModel('fractal')}
                    style={styles.modelChip}
                    icon="chart-line-variant"
                >分形预测</Chip>
            </ScrollView>
          </View>
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
                  mode="outlined"
                  onPress={() => handlePredict(stock.code)}
                  compact
                  icon="crystal-ball"
                  style={styles.predictBtn}
                  loading={predicting[stock.code]}
                >
                  AI 预测
                </Button>
              </View>
            </Card.Content>
          </Card>
        ))}
      </ScrollView>

      <Portal>
        <Dialog visible={predictionDialogVisible} onDismiss={() => setPredictionDialogVisible(false)}>
          <Dialog.Title>AI 预测结果</Dialog.Title>
          <Dialog.Content>
            {currentPrediction && (
              <ScrollView style={{ maxHeight: 400 }}>
                <Text variant="titleMedium" style={{ fontWeight: 'bold', marginBottom: 8 }}>
                  {currentPrediction.code} (置信度: {(currentPrediction.confidence * 100).toFixed(0)}%)
                </Text>
                
                {selectedModel === 'fractal' && currentFractalData && (
                    <FractalChart data={currentFractalData} />
                )}

                <Text variant="bodyMedium" style={{ lineHeight: 20 }}>
                  {currentPrediction.analysis}
                </Text>
                {currentPrediction.news_summary && (
                  <View style={styles.newsBox}>
                    <Text variant="bodySmall" style={styles.newsText}>
                      📰 {currentPrediction.news_summary}
                    </Text>
                  </View>
                )}
              </ScrollView>
            )}
          </Dialog.Content>
          <Dialog.Actions>
            <Button onPress={() => setPredictionDialogVisible(false)}>关闭</Button>
            <Button onPress={() => {
                setPredictionDialogVisible(false);
                // @ts-ignore
                navigation.navigate('Prediction', { code: currentPrediction?.code });
            }}>查看详情</Button>
          </Dialog.Actions>
        </Dialog>

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
  predictInputContainer: {
    paddingHorizontal: 16,
    paddingBottom: 8,
    backgroundColor: '#1E88E5',
  },
  predictInput: {
    backgroundColor: '#FFFFFF',
    height: 40,
  },
  modelSelectorContainer: {
    marginTop: 8,
    flexDirection: 'row',
  },
  modelChip: {
    marginRight: 8,
  },
});

export default HomeScreen;
