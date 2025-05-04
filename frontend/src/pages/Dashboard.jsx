import { useEffect, useState } from "react";
import {
  Box, Typography, Dialog, DialogTitle, DialogContent, TextField, DialogActions, Button, Tabs, Tab
} from "@mui/material";
import TopBar from "../components/TopBar";
import Sidebar from "../components/Sidebar";
import ProjectGroup from "../components/ProjectGroup";
import { useAuth } from "../context/AuthContext";
import api from "../services/api";
import AddIcon from "@mui/icons-material/Add";

export default function Dashboard() {
  const { role, userName } = useAuth();
  const [tabIndex, setTabIndex] = useState(0);
  const [channels, setChannels] = useState([]);
  const [clients, setClients] = useState([]);
  const [videos, setVideos] = useState([]);
  const [selectedFilter, setSelectedFilter] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [dialogLabel, setDialogLabel] = useState("");
  const [inputValue, setInputValue] = useState("");

  // Fetch sidebar data
  useEffect(() => {
    api.get("/sidebar-data").then((res) => {
      setChannels(res.data.channels);
      setClients(res.data.partners);
    });
  }, []);

  // Fetch projects based on filter/tab/role
  useEffect(() => {
    let query = "";
    if (role === "owner") {
      query =
        selectedFilter && tabIndex === 0
          ? `?channel_id=${selectedFilter.id}`
          : selectedFilter && tabIndex === 1
          ? `?editor_id=${selectedFilter.id}`
          : "";
    } else if (role === "editor") {
      query =
        selectedFilter && tabIndex === 0
          ? `?channel_id=${selectedFilter.id}`
          : "";
    }
    api.get(`/projects/recent${query}`).then((res) => {
      setVideos(res.data || []);
    });
  }, [tabIndex, selectedFilter, role]);

  // Dialog handlers
  const handleAddChannel = () => {
    setDialogLabel("Channel Name");
    setInputValue("");
    setOpenDialog(true);
  };
  const handleAddClient = () => {
    setDialogLabel("Client Email");
    setInputValue("");
    setOpenDialog(true);
  };
  const handleSubmitDialog = () => {
    alert(`${dialogLabel}: ${inputValue} added (dummy alert)`);
    setOpenDialog(false);
  };

  // Upload handler
  const handleUpload = (video) => {
    alert(`Upload to YouTube for video: ${video.title} (ID: ${video.id})`);
  };

  // Add this handler for OAuth
  const handleAddYouTubeAccount = () => {
    const token = localStorage.getItem("token");
    window.location.href = `http://localhost:8080/api/youtube/auth?state=${encodeURIComponent(token)}`;
  };

  // Group videos for display
  const groupedVideos = (videos || []).reduce((acc, video) => {
    let key;
    if (role === "owner") {
      key = tabIndex === 0 ? video.channel_name : video.editor_name;
    } else {
      key = video.channel_name;
    }
    (acc[key] = acc[key] || []).push(video);
    return acc;
  }, {});

  // Sidebar click handlers
  const handleChannelClick = (item) => {
    setSelectedFilter(item);
    setTabIndex(0);
  };
  const handleClientClick = (item) => {
    setSelectedFilter(item);
    setTabIndex(1);
  };

  // Tabs for owner/editor
  const tabs = role === "owner"
    ? [{ label: "Channels" }, { label: "Editors" }]
    : [{ label: "Channels" }, { label: "Clients" }];

  return (
    <Box display="flex" flexDirection="column" height="100vh">
      <TopBar onMenuClick={() => {}} />
      <Box display="flex" flex={1} minHeight={0}>
        {/* Sidebar */}
        <Box width="320px" minWidth="260px" borderRight={1} borderColor="divider" bgcolor="#f5f6fa">
          <Sidebar
            channels={channels}
            clients={clients}
            onChannelClick={handleChannelClick}
            onClientClick={role === "owner" ? handleClientClick : undefined}
            onAddChannel={handleAddChannel}
            onAddClient={role === "owner" ? handleAddClient : undefined}
          />
        </Box>
        {/* Main Content */}
        <Box flex={1} p={4} bgcolor="#fafbfc" overflow="auto">
          <Tabs value={tabIndex} onChange={(_, v) => setTabIndex(v)} sx={{ mb: 2 }}>
            {tabs.map((tab, idx) => (
              <Tab key={tab.label} label={tab.label} />
            ))}
          </Tabs>
          <Typography variant="h4" fontWeight={700} gutterBottom>
            Recent Projects
          </Typography>
          {Object.keys(groupedVideos).length === 0 ? (
            <Typography>No projects available.</Typography>
          ) : (
            Object.entries(groupedVideos).map(([group, vids]) => (
              <ProjectGroup key={group} group={group} videos={vids} onUpload={handleUpload} />
            ))
          )}
          {/* Add Dialog */}
          <Dialog open={openDialog} onClose={() => setOpenDialog(false)}>
            <DialogTitle>{dialogLabel}</DialogTitle>
            <DialogContent>
              <TextField
                fullWidth
                margin="dense"
                label={dialogLabel}
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
              />
            </DialogContent>
            <DialogActions>
              <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
              <Button onClick={handleSubmitDialog}>Submit</Button>
            </DialogActions>
          </Dialog>
        </Box>
      </Box>
    </Box>
  );
} 