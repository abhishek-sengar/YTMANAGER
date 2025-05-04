// src/components/TopBar.jsx
import React, { useState } from "react";
import {
  AppBar,
  Toolbar,
  IconButton,
  Typography,
  Avatar,
  Menu,
  MenuItem,
  Box,
} from "@mui/material";
import MenuIcon from "@mui/icons-material/Menu";
import EmojiEmotionsIcon from "@mui/icons-material/EmojiEmotions";
import { useAuth } from "../context/AuthContext";
import { useNavigate } from "react-router-dom";

export default function TopBar({ onMenuClick }) {
  const { userName = "User", logout } = useAuth?.() || {};
  const navigate = useNavigate?.() || (() => {});
  const [anchorEl, setAnchorEl] = useState(null);
  const open = Boolean(anchorEl);

  const handleAvatarClick = (e) => setAnchorEl(e.currentTarget);
  const handleMenuClose = () => setAnchorEl(null);
  const handleLogout = () => {
    logout?.();
    navigate?.("/login", { replace: true });
    handleMenuClose();
  };

  return (
    <AppBar position="static" color="default" elevation={1}>
      <Toolbar sx={{ justifyContent: "space-between" }}>
        {/* LEFT: Smiley Avatar and Greeting */}
        <Box display="flex" alignItems="center">
          <Avatar
            sx={{ bgcolor: "primary.main", cursor: "pointer", mr: 1, width: 40, height: 40 }}
            onClick={handleAvatarClick}
          >
            <EmojiEmotionsIcon fontSize="large" />
          </Avatar>
          <Menu
            anchorEl={anchorEl}
            open={open}
            onClose={handleMenuClose}
            anchorOrigin={{ vertical: "bottom", horizontal: "left" }}
            transformOrigin={{ vertical: "top", horizontal: "left" }}
          >
            <MenuItem onClick={handleLogout}>Logout</MenuItem>
          </Menu>
          <Typography variant="h6" component="div" sx={{ ml: 1, fontWeight: 600 }}>
            Hi, {userName}
          </Typography>
        </Box>
        {/* RIGHT: Hamburger Menu */}
        <IconButton edge="end" onClick={onMenuClick}>
          <MenuIcon fontSize="large" />
        </IconButton>
      </Toolbar>
    </AppBar>
  );
}
