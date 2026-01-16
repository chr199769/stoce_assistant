import React, { useState, useEffect } from 'react';
import { View, StyleSheet, ScrollView, RefreshControl } from 'react-native';
import { Text, Card, ActivityIndicator, Chip, useTheme } from 'react-native-paper';
import { analyzeMarket } from '../api/stock';
import { MarketAnalysisResponse } from '../types';

const MarketAnalysisScreen = () => {
  const theme = useTheme();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<MarketAnalysisResponse | null>(null);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await analyzeMarket({});
      setData(res);
    } catch (error) {
      console.error('Failed to fetch market analysis', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const renderContent = () => {
    if (!data) return (
       <Card style={styles.card}>
         <Card.Content>
           <Text>暂无盘前分析数据，请下拉刷新。</Text>
         </Card.Content>
       </Card>
    );

    return (
      <View style={styles.contentContainer}>
        {data.analysis_summary && (
           <Card style={styles.card}>
            <Card.Title title="分析总结" left={(props) => <Text {...props} style={{fontSize: 24}}>📝</Text>} />
            <Card.Content>
              <Text variant="bodyMedium">{data.analysis_summary}</Text>
            </Card.Content>
          </Card>
        )}

        {data.hot_stocks && data.hot_stocks.length > 0 && (
          <Card style={styles.card}>
            <Card.Title title="热门股票预测" left={(props) => <Text {...props} style={{fontSize: 24}}>🔥</Text>} />
            <Card.Content style={styles.chipContainer}>
              {data.hot_stocks.map((stock, index) => (
                <Chip key={index} style={styles.chip} mode="outlined">{stock}</Chip>
              ))}
            </Card.Content>
          </Card>
        )}

        {data.recommended_stocks && data.recommended_stocks.length > 0 && (
          <Card style={styles.card}>
            <Card.Title title="推荐关注" left={(props) => <Text {...props} style={{fontSize: 24}}>⭐</Text>} />
            <Card.Content>
              {data.recommended_stocks.map((stock, index) => (
                <View key={index} style={styles.recommendItem}>
                  <Text variant="bodyMedium">• {stock}</Text>
                </View>
              ))}
            </Card.Content>
          </Card>
        )}

        {data.opportunities && data.opportunities.length > 0 && (
           <Card style={styles.card}>
            <Card.Title title="机会展望" left={(props) => <Text {...props} style={{fontSize: 24}}>🚀</Text>} />
            <Card.Content>
              {data.opportunities.map((opp, index) => (
                <Text key={index} variant="bodyMedium" style={styles.oppItem}>• {opp}</Text>
              ))}
            </Card.Content>
          </Card>
        )}

        {data.risks && data.risks.length > 0 && (
          <Card style={styles.card}>
            <Card.Title title="风险提示" left={(props) => <Text {...props} style={{fontSize: 24}}>⚠️</Text>} />
            <Card.Content>
              {data.risks.map((risk, index) => (
                <Text key={index} variant="bodyMedium" style={styles.riskItem}>• {risk}</Text>
              ))}
            </Card.Content>
          </Card>
        )}
      </View>
    );
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text variant="headlineSmall" style={styles.headerTitle}>盘前分析</Text>
        <Text variant="bodySmall" style={styles.subHeader}>基于昨日盘面预测今日机会</Text>
      </View>

      <ScrollView
        contentContainerStyle={styles.scrollContent}
        refreshControl={
          <RefreshControl refreshing={loading} onRefresh={fetchData} />
        }
      >
        {loading && !data ? (
          <ActivityIndicator animating={true} size="large" style={styles.loading} />
        ) : (
          renderContent()
        )}
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
    padding: 16,
    backgroundColor: '#FFFFFF',
    elevation: 2,
  },
  headerTitle: {
    fontWeight: 'bold',
    color: '#1E88E5',
  },
  subHeader: {
    color: '#757575',
    marginTop: 4,
  },
  scrollContent: {
    padding: 16,
    paddingBottom: 32,
  },
  contentContainer: {
    gap: 16,
  },
  card: {
    marginBottom: 16,
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
  },
  loading: {
    marginTop: 40,
  },
  chipContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  chip: {
    marginRight: 8,
    marginBottom: 8,
  },
  recommendItem: {
    marginBottom: 8,
    padding: 8,
    backgroundColor: '#FFF9C4', // Light yellow for recommendations
    borderRadius: 4,
  },
  oppItem: {
    marginBottom: 4,
    color: '#2E7D32', // Green for opportunities
  },
  riskItem: {
    marginBottom: 4,
    color: '#D32F2F', // Red for risks
  },
});

export default MarketAnalysisScreen;
