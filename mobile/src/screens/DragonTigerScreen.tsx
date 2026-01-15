import React, { useEffect, useState } from 'react';
import { View, StyleSheet, FlatList, RefreshControl, ScrollView } from 'react-native';
import { Text, Card, ActivityIndicator, Appbar, Chip, Divider, Surface } from 'react-native-paper';
import { useNavigation } from '@react-navigation/native';
import { getDragonTigerList } from '../api/stock';
import { DragonTigerItem, DragonTigerSeat } from '../types';

const DragonTigerScreen = () => {
  const navigation = useNavigation();
  const [items, setItems] = useState<DragonTigerItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // Date handling: default to today (simplified, assumes backend handles empty date)
  // In a real app, adding a DatePicker would be good.
  const [date, setDate] = useState(''); 

  const fetchList = async () => {
    try {
      setError(null);
      const data = await getDragonTigerList(date);
      setItems(data.items || []);
    } catch (err) {
      setError('Failed to fetch Dragon Tiger List');
      console.error(err);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    fetchList();
  }, [date]);

  const onRefresh = () => {
    setRefreshing(true);
    fetchList();
  };

  const renderSeat = (seat: DragonTigerSeat, type: 'buy' | 'sell') => {
    return (
      <View style={styles.seatRow} key={seat.name}>
        <View style={{flex: 1}}>
           <Text variant="bodySmall" numberOfLines={1} style={{fontWeight: '500'}}>
             {seat.name}
           </Text>
           <View style={styles.tagsRow}>
             {seat.tags.map((tag, idx) => (
               <Chip key={idx} compact textStyle={{fontSize: 9}} style={styles.tag} mode="flat">
                 {tag}
               </Chip>
             ))}
           </View>
        </View>
        <Text variant="bodySmall" style={{color: type === 'buy' ? '#D32F2F' : '#388E3C'}}>
           {(seat.net_amt / 10000).toFixed(0)}万
        </Text>
      </View>
    );
  };

  const renderItem = ({ item }: { item: DragonTigerItem }) => {
    const isUp = item.change_percent >= 0;
    
    return (
      <Card style={styles.card}>
        <Card.Content>
          <View style={styles.cardHeader}>
            <View>
              <Text variant="titleMedium" style={{fontWeight: 'bold'}}>{item.name}</Text>
              <Text variant="bodySmall" style={{color: '#666'}}>{item.code}</Text>
            </View>
            <View style={{alignItems: 'flex-end'}}>
               <Text variant="titleMedium" style={{ color: isUp ? '#D32F2F' : '#388E3C', fontWeight: 'bold' }}>
                  {isUp ? '+' : ''}{item.change_percent.toFixed(2)}%
               </Text>
               <Text variant="bodySmall" style={{fontWeight: 'bold'}}>
                 Net: {(item.net_inflow / 10000).toFixed(0)}万
               </Text>
            </View>
          </View>
          
          <Text variant="bodySmall" style={styles.reason} numberOfLines={2}>
            Reason: {item.reason}
          </Text>
          
          <Divider style={{marginVertical: 8}} />
          
          {/* Seats Section */}
          <View style={styles.seatsContainer}>
             <View style={styles.seatColumn}>
                <Text variant="labelSmall" style={{color: '#D32F2F', marginBottom: 4}}>Top Buyers</Text>
                {item.buy_seats.length > 0 ? (
                  item.buy_seats.map(s => renderSeat(s, 'buy'))
                ) : (
                  <Text variant="bodySmall" style={{color: '#999'}}>No Data</Text>
                )}
             </View>
             <View style={{width: 10}} />
             <View style={styles.seatColumn}>
                <Text variant="labelSmall" style={{color: '#388E3C', marginBottom: 4}}>Top Sellers</Text>
                {item.sell_seats.length > 0 ? (
                  item.sell_seats.map(s => renderSeat(s, 'sell'))
                ) : (
                   <Text variant="bodySmall" style={{color: '#999'}}>No Data</Text>
                )}
             </View>
          </View>

        </Card.Content>
      </Card>
    );
  };

  return (
    <View style={styles.container}>
      <Appbar.Header>
        <Appbar.Content title="龙虎榜 (Dragon Tiger List)" subtitle={date || 'Today'} />
        <Appbar.Action icon="refresh" onPress={onRefresh} />
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
          data={items}
          renderItem={renderItem}
          keyExtractor={(item) => item.code}
          contentContainerStyle={styles.list}
          refreshControl={
            <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
          }
          ListEmptyComponent={
             <View style={styles.center}>
                <Text>No Dragon Tiger data available for today yet.</Text>
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
    padding: 20,
  },
  list: {
    padding: 10,
  },
  card: {
    marginBottom: 12,
    backgroundColor: 'white',
    elevation: 2,
  },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 5,
  },
  reason: {
    color: '#666',
    fontStyle: 'italic',
    fontSize: 11,
  },
  seatsContainer: {
    flexDirection: 'row',
  },
  seatColumn: {
    flex: 1,
  },
  seatRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 6,
  },
  tagsRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginTop: 2,
  },
  tag: {
    marginRight: 4,
    height: 18, 
    backgroundColor: '#E3F2FD',
  }
});

export default DragonTigerScreen;
