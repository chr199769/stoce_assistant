import React, { useEffect, useState } from 'react';
import { View, StyleSheet, FlatList, RefreshControl } from 'react-native';
import { Text, Card, ActivityIndicator, Appbar, DataTable } from 'react-native-paper';
import { useRoute, useNavigation, RouteProp } from '@react-navigation/native';
import { RootStackParamList } from '../navigation/AppNavigator';
import { getSectorStocks } from '../api/stock';
import { SectorStockItem } from '../types';

type SectorDetailRouteProp = RouteProp<RootStackParamList, 'SectorDetail'>;

const SectorDetailScreen = () => {
  const route = useRoute<SectorDetailRouteProp>();
  const navigation = useNavigation();
  const { sectorCode, sectorName } = route.params;

  const [stocks, setStocks] = useState<SectorStockItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchStocks = async () => {
    try {
      setError(null);
      const data = await getSectorStocks(sectorCode);
      setStocks(data.stocks || []);
    } catch (err) {
      setError('Failed to fetch sector stocks');
      console.error(err);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    fetchStocks();
  }, [sectorCode]);

  const onRefresh = () => {
    setRefreshing(true);
    fetchStocks();
  };

  const renderItem = ({ item }: { item: SectorStockItem }) => {
    const isUp = item.change_percent >= 0;
    const color = isUp ? '#D32F2F' : '#388E3C'; // Red for up, Green for down (CN market style)
    
    // Determine Market
    let market = 'SZ';
    if (item.code.startsWith('6')) market = 'SH';
    else if (item.code.startsWith('8') || item.code.startsWith('4')) market = 'BJ';

    return (
      <Card style={styles.card} onPress={() => {
        // Navigate to Prediction with code
        // @ts-ignore
        navigation.navigate('Prediction', { code: item.code });
      }}>
        <Card.Content style={styles.cardContent}>
          <View style={styles.leftCol}>
            <Text variant="titleMedium" style={styles.stockName}>{item.name}</Text>
            <View style={styles.codeBadge}>
                <Text style={styles.codeText}>{market}{item.code}</Text>
            </View>
          </View>
          
          <View style={styles.midCol}>
            <Text variant="bodyMedium">Price: {item.price.toFixed(2)}</Text>
            <Text variant="bodySmall" style={{ color: '#757575' }}>
              Vol: {(item.volume / 10000).toFixed(1)}万
            </Text>
          </View>

          <View style={styles.rightCol}>
            <Text variant="titleMedium" style={{ color, fontWeight: 'bold' }}>
              {isUp ? '+' : ''}{item.change_percent.toFixed(2)}%
            </Text>
            <Text variant="bodySmall" style={{ color: '#757575' }}>
              Amt: {(item.amount / 100000000).toFixed(1)}亿
            </Text>
          </View>
        </Card.Content>
      </Card>
    );
  };

  return (
    <View style={styles.container}>
      <Appbar.Header>
        <Appbar.BackAction onPress={() => navigation.goBack()} />
        <Appbar.Content title={`${sectorName} - Sector Leaders`} subtitle={`Code: ${sectorCode}`} />
      </Appbar.Header>

      {loading && !refreshing ? (
        <View style={styles.center}>
          <ActivityIndicator size="large" />
        </View>
      ) : error ? (
        <View style={styles.center}>
          <Text style={{ color: 'red' }}>{error}</Text>
        </View>
      ) : (
        <FlatList
          data={stocks}
          renderItem={renderItem}
          keyExtractor={(item) => item.code}
          contentContainerStyle={styles.list}
          refreshControl={
            <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
          }
          ListHeaderComponent={
            <View style={styles.header}>
              <Text variant="bodySmall" style={{color: '#666'}}>
                * Stocks sorted by influence (Amount & Market Cap). Top 5 leaders highlighted.
              </Text>
            </View>
          }
        />
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  center: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  list: {
    padding: 10,
  },
  header: {
    paddingBottom: 10,
    paddingHorizontal: 5,
  },
  card: {
    marginBottom: 10,
    backgroundColor: 'white',
  },
  cardContent: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  leftCol: {
    flex: 2,
  },
  midCol: {
    flex: 2,
    alignItems: 'flex-end',
    paddingRight: 10,
  },
  rightCol: {
    flex: 1.5,
    alignItems: 'flex-end',
  },
  stockName: {
    fontWeight: 'bold',
  },
  codeBadge: {
    backgroundColor: '#E0E0E0',
    paddingHorizontal: 4,
    paddingVertical: 2,
    borderRadius: 4,
    alignSelf: 'flex-start',
    marginTop: 4,
  },
  codeText: {
    fontSize: 10,
    color: '#424242',
  },
});

export default SectorDetailScreen;
