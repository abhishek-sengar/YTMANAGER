// src/pages/Signup.jsx
import { useState } from "react";
import {
  Container,
  Box,
  Typography,
  TextField,
  Button,
  Link,
  Paper,
  ToggleButtonGroup,
  ToggleButton,
  Divider,
  Stack,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import GoogleIcon from "@mui/icons-material/Google";
import FacebookIcon from "@mui/icons-material/Facebook";
import api from "../services/api";

export default function Signup() {
  const [form, setForm] = useState({
    name: "",
    email: "",
    password: "",
    role: "editor",
  });
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleRoleChange = (_, newRole) => {
    if (newRole !== null) {
      setForm({ ...form, role: newRole });
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    try {
      await api.post("/signup", form);
      navigate("/login");
    } catch (err) {
      setError(err.response?.data?.error || "Signup failed");
    }
  };

  return (
    <Container maxWidth="sm">
      <Paper elevation={3} sx={{ p: 4, mt: 8 }}>
        <Typography variant="h4" align="center" gutterBottom>
          Sign Up
        </Typography>

        <Box component="form" onSubmit={handleSubmit} noValidate>
          <TextField
            fullWidth
            margin="normal"
            required
            label="Name"
            name="name"
            value={form.name}
            onChange={handleChange}
          />
          <TextField
            fullWidth
            margin="normal"
            required
            label="Email"
            name="email"
            type="email"
            value={form.email}
            onChange={handleChange}
          />
          <TextField
            fullWidth
            margin="normal"
            required
            label="Password"
            name="password"
            type="password"
            value={form.password}
            onChange={handleChange}
          />

          <Typography variant="subtitle1" mt={2}>
            I am signing up as:
          </Typography>
          <ToggleButtonGroup
            color="primary"
            exclusive
            fullWidth
            value={form.role}
            onChange={handleRoleChange}
            sx={{ mt: 1 }}
          >
            <ToggleButton value="editor">Editor</ToggleButton>
            <ToggleButton value="owner">Owner</ToggleButton>
          </ToggleButtonGroup>

          {error && (
            <Typography color="error" variant="body2" mt={2}>
              {error}
            </Typography>
          )}

          <Button type="submit" fullWidth variant="contained" sx={{ mt: 3 }}>
            Sign Up
          </Button>

          <Typography align="center" variant="body2" mt={2}>
            Already have an account?{" "}
            <Link href="/login" underline="hover">
              Login
            </Link>
          </Typography>
        </Box>

        <Divider sx={{ my: 3 }}>or</Divider>

        <Stack spacing={2}>
          <Button
            fullWidth
            variant="outlined"
            startIcon={<GoogleIcon />}
            onClick={() => alert("Google login not implemented")}
          >
            Sign up with Google
          </Button>
          <Button
            fullWidth
            variant="outlined"
            startIcon={<FacebookIcon />}
            onClick={() => alert("Facebook login not implemented")}
          >
            Sign up with Facebook
          </Button>
        </Stack>
      </Paper>
    </Container>
  );
}
