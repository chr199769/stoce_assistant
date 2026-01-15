import React, { useEffect, useState } from 'react';
import { View, StyleSheet, ScrollView, Dimensions, TouchableOpacity } from 'react-native';
import { Appbar, Card, Text, Divider, Chip, Button, ActivityIndicator } from 'react-native-paper';
import { useNavigation } from '@react-navigation/native';
import { getRealtime, marketReview } from '../api/stock';
import { RealtimeResponse, MarketReviewResponse } from '../types';
import { PieChart } from 'react-native-chart-kit';

const SummaryScreen = () => {
  const navigation = useNavigation();
  const [indices, setIndices] = useState<RealtimeResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [review, setReview] = useState<MarketReviewResponse | null>(null);
  const [reviewLoading, setReviewLoading] = useState(false);

  // Mock sector data
  const sectorData = [
    { name: '科技', population: 35, color: '#F44336', legendFontColor: '#7F7F7F', legendFontSize: 12 },
    { name: '金融', population: 25, color: '#1E88E5', legendFontColor: '#7F7F7F', legendFontSize: 12 },
    { name: '消费', population: 20, color: '#FF9800', legendFontColor: '#7F7F7F', legendFontSize: 12 },
    { name: '医疗', population: 15, color: '#4CAF50', legendFontColor: '#7F7F7F', legendFontSize: 12 },
    { name: '其他', population: 5, color: '#9E9E9E', legendFontColor: '#7F7F7F', legendFontSize: 12 },
  ];

  const fetchIndices = async () => {
    setLoading(true);
    try {
      // Fetch major indices
      const codes = ['sh000001', 'sz399001', 'sz399006']; // ShangZheng, ShenZheng, ChuangYe
      const promises = codes.map(c => getRealtime(c));
      const results = await Promise.all(promises);
      setIndices(results);
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const fetchReview = async () => {
    setReviewLoading(true);
    try {
      const today = new Date().toISOString().split('T')[0];
      const res = await marketReview({ date: today });
      setReview(res);
      // Parse Sector Analysis from review text if possible, or just use static for now.
      // In a real app, the review response would have structured sector data.
    } catch (error) {
      console.error("Failed to fetch market review:", error);
    } finally {
      setReviewLoading(false);
    }
  };

  const handleSectorPress = (sectorName: string) => {
    // Map common names to codes (Simplified mapping for demo)
    // In a real app, you'd get code from the API response
    const codeMap: Record<string, string> = {
      '半导体': 'BK1036',
      '新能源车': 'BK0900',
      '人工智能': 'BK0985',
      '科技': 'BK0696',
      '金融': 'BK0475'
    };

    // Default to a tech sector if not found
    const code = codeMap[sectorName] || 'BK0800';

    // @ts-ignore
    navigation.navigate('SectorDetail', { sectorCode: code, sectorName });
  };

  useEffect(() => {
    fetchIndices();
    fetchReview();
  }, []);

  const getColor = (change: number) => {
    if (change > 0) return '#F44336';
    if (change < 0) return '#4CAF50';
    return '#333333';
  };

  return (
    <View style={styles.container}>
      <Appbar.Header style={styles.header}>
        <Appbar.Content title="大盘总结" titleStyle={styles.headerTitle} />
      </Appbar.Header>

      <ScrollView contentContainerStyle={styles.content}>
        <Text variant="titleMedium" style={styles.sectionTitle}>主要指数</Text>
        <View style={styles.indicesContainer}>
          {indices.map((index) => (
            <Card key={index.code} style={styles.indexCard}>
              <Card.Content style={styles.indexCardContent}>
                <Text variant="bodyMedium" style={{ fontWeight: 'bold' }}>{index.name}</Text>
                <Text variant="titleMedium" style={{ color: getColor(index.change_percent) }}>
                  {index.current_price.toFixed(2)}
                </Text>
                <Text variant="bodySmall" style={{ color: getColor(index.change_percent) }}>
                  {index.change_percent > 0 ? '+' : ''}{index.change_percent.toFixed(2)}%
                </Text>
              </Card.Content>
            </Card>
          ))}
        </View>

        <Divider style={styles.divider} />

        <Text variant="titleMedium" style={styles.sectionTitle}>AI 市场复盘</Text>
        <Card style={styles.reviewCard}>
          <Card.Content>
            {reviewLoading ? (
              <ActivityIndicator animating={true} color="#1E88E5" />
            ) : review ? (
              <>
                <Text variant="bodyMedium" style={styles.reviewText}>{review.summary}</Text>
                <View style={styles.confidenceContainer}>
                  <Text variant="bodySmall" style={styles.confidenceLabel}>置信度:</Text>
                  <Text variant="bodySmall" style={styles.confidenceValue}>{(review.confidence * 100).toFixed(0)}%</Text>
                </View>
                <Text variant="bodySmall" style={styles.dateText}>日期: {review.date}</Text>
              </>
            ) : (
              <Button mode="outlined" onPress={fetchReview}>加载复盘</Button>
            )}
          </Card.Content>
        </Card>

        <Divider style={styles.divider} />

        <Text variant="titleMedium" style={styles.sectionTitle}>板块涨跌分布</Text>
        <Card style={styles.chartCard}>
          <Card.Content>
            <PieChart
              data={sectorData}
              width={Dimensions.get('window').width - 64}
              height={220}
              chartConfig={{
                backgroundColor: '#ffffff',
                backgroundGradientFrom: '#ffffff',
                backgroundGradientTo: '#ffffff',
                color: (opacity = 1) => `rgba(0, 0, 0, ${opacity})`,
              }}
              accessor="population"
              backgroundColor="transparent"
              paddingLeft="15"
              absolute
            />
          </Card.Content>
        </Card>

        <Text variant="titleMedium" style={styles.sectionTitle}>今日热点 (点击查看详情)</Text>
        <View style={styles.chipContainer}>
          <Chip icon="fire" style={styles.chip} textStyle={{ color: '#fff' }} onPress={() => handleSectorPress('半导体')}>半导体</Chip>
          <Chip icon="fire" style={styles.chip} textStyle={{ color: '#fff' }} onPress={() => handleSectorPress('新能源车')}>新能源车</Chip>
          <Chip icon="trending-up" style={styles.chip} textStyle={{ color: '#fff' }} onPress={() => handleSectorPress('人工智能')}>人工智能</Chip>
        </View>

      </ScrollView>
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
  content: {
    padding: 16,
  },
  sectionTitle: {
    fontWeight: 'bold',
    marginBottom: 12,
    marginTop: 8,
    color: '#333',
  },
  indicesContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    flexWrap: 'wrap',
  },
  indexCard: {
    width: '31%',
    backgroundColor: '#FFFFFF',
    marginBottom: 8,
  },
  indexCardContent: {
    alignItems: 'center',
    paddingHorizontal: 4,
    paddingVertical: 8,
  },
  reviewCard: {
    backgroundColor: '#FFFFFF',
    marginBottom: 8,
  },
  reviewText: {
    lineHeight: 24,
    color: '#333',
    marginBottom: 8,
  },
  confidenceContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginTop: 8,
  },
  confidenceLabel: {
    color: '#666',
    marginRight: 4,
  },
  confidenceValue: {
    color: '#1E88E5',
    fontWeight: 'bold',
  },
  dateText: {
    color: '#999',
    marginTop: 4,
    textAlign: 'right',
  },
  divider: {
    marginVertical: 16,
  },
  chartCard: {
    backgroundColor: '#FFFFFF',
    alignItems: 'center',
    marginBottom: 16,
  },
  chipContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
  },
  chip: {
    marginRight: 8,
    marginBottom: 8,
    backgroundColor: '#F44336', // Red for hot
  },
});

export default SummaryScreen;
