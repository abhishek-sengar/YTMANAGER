import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Typography,
  TextField,
  Button,
  Container,
  Paper,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from "@mui/material";
import { useAuth } from "../context/AuthContext";
import api from "../services/api";

export default function CreateProject() {
  const navigate = useNavigate();
  const { role } = useAuth();
  const [form, setForm] = useState({
    title: "",
    description: "",
    channel_id: "",
    status: "draft",
  });
  const [channels, setChannels] = useState([]);
  const [error, setError] = useState("");

  // Fetch available channels
  useState(() => {
    api.get("/channels").then((res) => {
      setChannels(res.data);
    });
  }, []);

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await api.post("/projects", form);
      navigate("/dashboard");
    } catch (err) {
      setError(err.response?.data?.error || "Failed to create project");
    }
  };

  return (
    <Container maxWidth="md">
      <Paper elevation={3} sx={{ p: 4, mt: 4 }}>
        <Typography variant="h4" gutterBottom>
          Create New Project
        </Typography>

        <Box component="form" onSubmit={handleSubmit} sx={{ mt: 3 }}>
          <TextField
            fullWidth
            required
            label="Title"
            name="title"
            value={form.title}
            onChange={handleChange}
            margin="normal"
          />

          <TextField
            fullWidth
            required
            label="Description"
            name="description"
            value={form.description}
            onChange={handleChange}
            margin="normal"
            multiline
            rows={4}
          />

          <FormControl fullWidth margin="normal">
            <InputLabel>Channel</InputLabel>
            <Select
              name="channel_id"
              value={form.channel_id}
              onChange={handleChange}
              required
            >
              {channels.map((channel) => (
                <MenuItem key={channel.id} value={channel.id}>
                  {channel.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          {error && (
            <Typography color="error" sx={{ mt: 2 }}>
              {error}
            </Typography>
          )}

          <Box sx={{ mt: 3, display: "flex", gap: 2 }}>
            <Button
              variant="contained"
              type="submit"
              disabled={role !== "editor"}
            >
              Create Project
            </Button>
            <Button variant="outlined" onClick={() => navigate("/dashboard")}>
              Cancel
            </Button>
          </Box>
        </Box>
      </Paper>
    </Container>
  );
} 