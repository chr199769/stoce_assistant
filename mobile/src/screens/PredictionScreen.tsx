import React, { useState, useEffect } from 'react';
import { View, StyleSheet, ScrollView, ActivityIndicator } from 'react-native';
import { Appbar, TextInput, Button, Card, Text, ProgressBar, HelperText, Chip } from 'react-native-paper';
import { getPrediction } from '../api/stock';
import { PredictionResponse } from '../types';
import { useRoute } from '@react-navigation/native';

const PredictionScreen = () => {
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<PredictionResponse | null>(null);
  const [error, setError] = useState('');

  const route = useRoute();

  useEffect(() => {
    // @ts-ignore
    if (route.params?.code) {
      // @ts-ignore
      setCode(route.params.code);
      // @ts-ignore
      handlePredict(route.params.code);
    }
  }, [route.params]);

  const handlePredict = async (searchCode: string = code) => {
    if (!searchCode) return;
    setLoading(true);
    setError('');
    setResult(null);
    try {
      const data = await getPrediction({
        code: searchCode,
        days: 3,
        include_news: true,
        model: 'glm-4.6v-flash', // Default model
      });
      setResult(data);
    } catch (err) {
      setError('获取预测失败，请检查股票代码或网络连接');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.container}>
      <Appbar.Header style={styles.header}>
        <Appbar.Content title="个股预测" titleStyle={styles.headerTitle} />
      </Appbar.Header>

      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.searchContainer}>
          <TextInput
            mode="outlined"
            label="股票代码 (如 sh600519)"
            value={code}
            onChangeText={setCode}
            style={styles.input}
            right={<TextInput.Icon icon="magnify" onPress={() => handlePredict(code)} />}
          />
          <Button mode="contained" onPress={() => handlePredict(code)} loading={loading} style={styles.button}>
            开始预测
          </Button>
        </View>

        {error ? <HelperText type="error">{error}</HelperText> : null}

        {loading && <ActivityIndicator size="large" color="#1E88E5" style={{ marginTop: 20 }} />}

        {result && (
          <View>
            <Card style={styles.card}>
              <Card.Title title="预测结果分析" />
              <Card.Content>
                <Text variant="titleLarge" style={styles.stockTitle}>{result.code}</Text>
                
                <View style={styles.confidenceContainer}>
                  <Text variant="bodyMedium">置信度: {(result.confidence * 100).toFixed(1)}%</Text>
                  <ProgressBar progress={result.confidence} color="#1E88E5" style={styles.progressBar} />
                </View>

                <Text variant="titleMedium" style={styles.sectionTitle}>走势分析</Text>
                <Text variant="bodyMedium" style={styles.analysisText}>{result.analysis}</Text>

                <Text variant="titleMedium" style={styles.sectionTitle}>新闻摘要</Text>
                <Text variant="bodySmall" style={styles.newsText}>{result.news_summary}</Text>
              </Card.Content>
            </Card>
          </View>
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
    backgroundColor: '#1E88E5',
  },
  headerTitle: {
    color: '#FFFFFF',
    fontWeight: 'bold',
  },
  content: {
    padding: 16,
  },
  searchContainer: {
    marginBottom: 16,
  },
  input: {
    backgroundColor: '#FFFFFF',
    marginBottom: 12,
  },
  button: {
    backgroundColor: '#1E88E5',
  },
  card: {
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    elevation: 2,
    marginBottom: 16,
  },
  stockTitle: {
    fontWeight: 'bold',
    marginBottom: 8,
    color: '#333',
  },
  confidenceContainer: {
    marginBottom: 16,
  },
  progressBar: {
    height: 8,
    borderRadius: 4,
    marginTop: 4,
  },
  sectionTitle: {
    fontWeight: 'bold',
    marginTop: 12,
    marginBottom: 4,
    color: '#1E88E5',
  },
  analysisText: {
    lineHeight: 22,
    color: '#424242',
  },
  newsText: {
    lineHeight: 20,
    color: '#616161',
  },
});

export default PredictionScreen;
