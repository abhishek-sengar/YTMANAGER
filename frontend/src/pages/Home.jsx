import { Box, Typography, Button, Container } from "@mui/material";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export default function Home() {
  const navigate = useNavigate();
  const { token } = useAuth();

  return (
    <Container maxWidth="md">
      <Box
        sx={{
          minHeight: "80vh",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          textAlign: "center",
          gap: 4,
        }}
      >
        <Typography variant="h2" component="h1" gutterBottom>
          Welcome to YouTube Manager
        </Typography>
        <Typography variant="h5" color="text.secondary" paragraph>
          Manage your YouTube channels and content in one place
        </Typography>
        {token ? (
          <Button
            variant="contained"
            size="large"
            onClick={() => navigate("/dashboard")}
          >
            Go to Dashboard
          </Button>
        ) : (
          <Box sx={{ display: "flex", gap: 2 }}>
            <Button
              variant="contained"
              size="large"
              onClick={() => navigate("/login")}
            >
              Login
            </Button>
            <Button
              variant="outlined"
              size="large"
              onClick={() => navigate("/signup")}
            >
              Sign Up
            </Button>
          </Box>
        )}
      </Box>
    </Container>
  );
} 