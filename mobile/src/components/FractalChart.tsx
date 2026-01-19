import React from 'react';
import { View, Text, StyleSheet } from 'react-native';

interface Props {
  symbol: string;
}

const FractalChart: React.FC<Props> = ({ symbol }) => {
  return (
    <View style={styles.container}>
      <Text style={styles.text}>Fractal Chart for {symbol}</Text>
      <Text style={styles.subtext}>Pattern Matching Visualization (Coming Soon)</Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    height: 200,
    backgroundColor: '#f8f9fa',
    justifyContent: 'center',
    alignItems: 'center',
    margin: 16,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#eee',
  },
  text: { fontSize: 16, fontWeight: 'bold', color: '#555' },
  subtext: { fontSize: 12, color: '#999', marginTop: 8 },
});

export default FractalChart;
