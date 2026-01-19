import React from 'react';
import { View, Text, FlatList, TouchableOpacity, StyleSheet } from 'react-native';

interface Stock {
  symbol: string;
  name: string;
  price: number;
  change: number;
}

interface Props {
  data: Stock[];
  onPress: (stock: Stock) => void;
}

const StockList: React.FC<Props> = ({ data, onPress }) => {
  return (
    <FlatList
      data={data}
      keyExtractor={(item) => item.symbol}
      renderItem={({ item }) => (
        <TouchableOpacity style={styles.item} onPress={() => onPress(item)}>
          <View>
            <Text style={styles.symbol}>{item.name}</Text>
            <Text style={styles.code}>{item.symbol}</Text>
          </View>
          <View style={styles.right}>
            <Text style={styles.price}>{item.price.toFixed(2)}</Text>
            <Text style={[styles.change, { color: item.change >= 0 ? '#e74c3c' : '#2ecc71' }]}>
              {item.change >= 0 ? '+' : ''}{item.change.toFixed(2)}%
            </Text>
          </View>
        </TouchableOpacity>
      )}
    />
  );
};

const styles = StyleSheet.create({
  item: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
    backgroundColor: 'white',
  },
  symbol: { fontSize: 16, fontWeight: 'bold', color: '#333' },
  code: { color: '#999', fontSize: 12, marginTop: 4 },
  right: { alignItems: 'flex-end' },
  price: { fontSize: 16, fontWeight: 'bold', color: '#333' },
  change: { fontSize: 14, fontWeight: '500', marginTop: 4 },
});

export default StockList;
