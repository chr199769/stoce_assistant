import React, { useState } from 'react';
import { StyleSheet, Alert } from 'react-native';
import { Button, TextInput, Title, Text, Surface } from 'react-native-paper';
import { useAuth } from '../context/AuthContext';
import { SafeAreaView } from 'react-native-safe-area-context';

const LoginScreen = () => {
  const [username, setUsername] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();

  const handleLogin = async () => {
    if (!username.trim()) {
      Alert.alert('Error', 'Please enter a username');
      return;
    }

    setLoading(true);
    try {
      await login(username);
      // Navigation will be handled by the navigator based on auth state
    } catch {
      Alert.alert('Login Failed', 'Could not connect to server or create user.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <Surface style={styles.content} elevation={2}>
        <Title style={styles.title}>Stock Assistant</Title>
        <Text style={styles.subtitle}>Enter your username to continue</Text>
        
        <TextInput
          label="Username"
          value={username}
          onChangeText={setUsername}
          mode="outlined"
          style={styles.input}
          autoCapitalize="none"
        />

        <Button 
          mode="contained" 
          onPress={handleLogin} 
          loading={loading}
          disabled={loading}
          style={styles.button}
        >
          Login / Register
        </Button>
      </Surface>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    padding: 20,
    backgroundColor: '#f5f5f5',
  },
  content: {
    padding: 20,
    borderRadius: 8,
    alignItems: 'center',
    backgroundColor: 'white',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 8,
    color: '#1E88E5',
  },
  subtitle: {
    marginBottom: 24,
    color: '#666',
  },
  input: {
    width: '100%',
    marginBottom: 16,
  },
  button: {
    width: '100%',
    marginTop: 8,
    paddingVertical: 6,
  },
});

export default LoginScreen;
