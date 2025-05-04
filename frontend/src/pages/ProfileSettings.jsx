import { useState, useEffect } from "react";
import {
  Box,
  Typography,
  TextField,
  Button,
  Container,
  Paper,
  Divider,
  Alert,
} from "@mui/material";
import { useAuth } from "../context/AuthContext";
import api from "../services/api";

export default function ProfileSettings() {
  const { userName, token } = useAuth();
  const [form, setForm] = useState({
    name: "",
    email: "",
    currentPassword: "",
    newPassword: "",
    confirmPassword: "",
  });
  const [message, setMessage] = useState({ type: "", text: "" });

  useEffect(() => {
    // Fetch user profile data
    api.get("/profile").then((res) => {
      setForm((prev) => ({
        ...prev,
        name: res.data.name || "",
        email: res.data.email || "",
      }));
    });
  }, []);

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleProfileUpdate = async (e) => {
    e.preventDefault();
    try {
      await api.put("/profile", {
        name: form.name,
        email: form.email,
      });
      setMessage({ type: "success", text: "Profile updated successfully" });
    } catch (err) {
      setMessage({
        type: "error",
        text: err.response?.data?.error || "Failed to update profile",
      });
    }
  };

  const handlePasswordUpdate = async (e) => {
    e.preventDefault();
    if (form.newPassword !== form.confirmPassword) {
      setMessage({ type: "error", text: "Passwords do not match" });
      return;
    }

    try {
      await api.put("/profile/password", {
        currentPassword: form.currentPassword,
        newPassword: form.newPassword,
      });
      setMessage({ type: "success", text: "Password updated successfully" });
      setForm((prev) => ({
        ...prev,
        currentPassword: "",
        newPassword: "",
        confirmPassword: "",
      }));
    } catch (err) {
      setMessage({
        type: "error",
        text: err.response?.data?.error || "Failed to update password",
      });
    }
  };

  return (
    <Container maxWidth="md">
      <Paper elevation={3} sx={{ p: 4, mt: 4 }}>
        <Typography variant="h4" gutterBottom>
          Profile Settings
        </Typography>

        {message.text && (
          <Alert severity={message.type} sx={{ mb: 3 }}>
            {message.text}
          </Alert>
        )}

        {/* Profile Update Form */}
        <Box component="form" onSubmit={handleProfileUpdate} sx={{ mb: 4 }}>
          <Typography variant="h6" gutterBottom>
            Update Profile
          </Typography>
          <TextField
            fullWidth
            label="Name"
            name="name"
            value={form.name}
            onChange={handleChange}
            margin="normal"
          />
          <TextField
            fullWidth
            label="Email"
            name="email"
            type="email"
            value={form.email}
            onChange={handleChange}
            margin="normal"
          />
          <Button
            type="submit"
            variant="contained"
            sx={{ mt: 2 }}
          >
            Update Profile
          </Button>
        </Box>

        <Divider sx={{ my: 4 }} />

        {/* Password Update Form */}
        <Box component="form" onSubmit={handlePasswordUpdate}>
          <Typography variant="h6" gutterBottom>
            Change Password
          </Typography>
          <TextField
            fullWidth
            label="Current Password"
            name="currentPassword"
            type="password"
            value={form.currentPassword}
            onChange={handleChange}
            margin="normal"
          />
          <TextField
            fullWidth
            label="New Password"
            name="newPassword"
            type="password"
            value={form.newPassword}
            onChange={handleChange}
            margin="normal"
          />
          <TextField
            fullWidth
            label="Confirm New Password"
            name="confirmPassword"
            type="password"
            value={form.confirmPassword}
            onChange={handleChange}
            margin="normal"
          />
          <Button
            type="submit"
            variant="contained"
            sx={{ mt: 2 }}
          >
            Update Password
          </Button>
        </Box>
      </Paper>
    </Container>
  );
} 