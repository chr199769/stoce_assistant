import React from 'react';
import { View, Text, StyleSheet, Dimensions } from 'react-native';
import { LineChart } from 'react-native-chart-kit';

export interface FractalData {
  query: number[];
  match: number[];
  projection: number[];
  match_date: string;
}

interface Props {
  data: FractalData | null;
}

const FractalChart: React.FC<Props> = ({ data }) => {
  if (!data || !data.query || data.query.length === 0) return null;

  const queryData = data.query;
  const projectionData = data.projection;
  
  // Combine data: Query + Projection
  const combinedData = [...queryData, ...projectionData];
  
  // Labels: Mark 'Now' at the transition point
  const labels = combinedData.map((_, i) => {
    if (i === 0) return '起点';
    if (i === queryData.length - 1) return '当前';
    if (i === combinedData.length - 1) return '终点';
    return '';
  });

  const chartData = {
    labels: labels,
    datasets: [
      {
        data: combinedData,
        color: (opacity = 1) => `rgba(30, 136, 229, ${opacity})`, // Main Blue Line
        strokeWidth: 2
      },
      // We can optionally add the "Matched History" line for comparison
      // But we need to scale it to the current price range first.
      // For now, let's keep it simple.
    ],
    legend: ["分形走势推演"]
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>分形预测走势图</Text>
      <Text style={styles.subtitle}>匹配历史时段: {data.match_date}</Text>
      <LineChart
        data={chartData}
        width={Dimensions.get("window").width - 48} // Padding adjustment
        height={220}
        yAxisLabel=""
        yAxisSuffix=""
        yAxisInterval={1}
        chartConfig={{
          backgroundColor: "#ffffff",
          backgroundGradientFrom: "#ffffff",
          backgroundGradientTo: "#ffffff",
          decimalPlaces: 2,
          color: (opacity = 1) => `rgba(30, 136, 229, ${opacity})`,
          labelColor: (opacity = 1) => `rgba(100, 100, 100, ${opacity})`,
          style: {
            borderRadius: 16
          },
          propsForDots: {
            r: "3",
            strokeWidth: "1",
            stroke: "#1E88E5"
          },
          propsForBackgroundLines: {
            strokeDasharray: "", // solid background lines
            stroke: "#eee"
          }
        }}
        bezier
        style={{
          marginVertical: 8,
          borderRadius: 8,
        }}
        // Highlight the 'Now' point
        getDotColor={(dataPoint, index) => index === queryData.length - 1 ? '#D32F2F' : 'transparent'}
        renderDotContent={({x, y, index}) => {
             if (index === queryData.length - 1) {
                return (
                    <Text
                        key={index}
                        style={{
                        position: 'absolute',
                        top: y - 25,
                        left: x - 10,
                        fontSize: 10,
                        color: '#D32F2F',
                        fontWeight: 'bold'
                        }}>
                        当前
                    </Text>
                )
            }
            return null;
        }}
      />
      <View style={styles.legendContainer}>
        <View style={styles.legendItem}>
             <View style={[styles.dot, {backgroundColor: '#D32F2F'}]} />
             <Text style={styles.legendText}>当前时点</Text>
        </View>
        <Text style={styles.legendText}> • </Text>
        <Text style={styles.legendText}>左侧: 历史走势 | 右侧: 未来推演</Text>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#fff',
    padding: 16,
    borderRadius: 8,
    marginVertical: 16,
    elevation: 2,
    alignItems: 'center'
  },
  title: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 12,
    color: '#666',
    marginBottom: 12,
  },
  legendContainer: {
    flexDirection: 'row',
    marginTop: 8,
    alignItems: 'center'
  },
  legendItem: {
    flexDirection: 'row',
    alignItems: 'center'
  },
  dot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: 4
  },
  legendText: {
    fontSize: 12,
    color: '#666'
  }
});

export default FractalChart;
