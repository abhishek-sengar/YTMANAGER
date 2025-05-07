import {
  Box,
  Typography,
  IconButton,
  InputBase,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  Avatar,
  Divider,
  Paper,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Checkbox,
  CircularProgress,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import SearchIcon from "@mui/icons-material/Search";
import VideoUploadIcon from "@mui/icons-material/VideoLibrary";
import { useState, useEffect } from "react";
import { useAuth } from "../context/AuthContext";
import api from "../services/api";
import { toast } from "react-hot-toast";
import VideoUpload from "./VideoUpload";

export default function Sidebar({
  channels,
  clients,
  onChannelClick,
  onClientClick,
  onAddChannel,
  onAddClient,
}) {
  const [search, setSearch] = useState({ channels: "", clients: "" });
  const { role } = useAuth();
  const [addChannelDialogOpen, setAddChannelDialogOpen] = useState(false);
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false);

  const handleAddChannel = () => {
    setAddChannelDialogOpen(true);
  };

  const handleAddYouTubeAccount = () => {
    const token = localStorage.getItem("token");
    if (!token) {
      window.location.href = "/login";
      return;
    }
    window.location.href = `http://localhost:8080/api/youtube/auth?state=${encodeURIComponent(token)}`;
  };

  const AddChannelDialog = () => {
    const [allChannels, setAllChannels] = useState([]);
    const [selectedChannelIds, setSelectedChannelIds] = useState([]);
    const [isLoading, setIsLoading] = useState(false);

    useEffect(() => {
      const fetchChannels = async () => {
        setIsLoading(true);
        try {
          // First fetch unattached channels
          const res = await api.get("/api/youtube/unattached-channels");
          const unattachedChannels = res.data.channels || [];
          
          // Then fetch currently attached channels from sidebar data
          const sidebarRes = await api.get("/sidebar-data");
          const attachedChannels = sidebarRes.data.channels || [];
          
          // Combine both lists
          const combinedChannels = [
            ...attachedChannels.map(ch => ({ 
              ...ch, 
              is_attached: true,
              youtube_account_id: ch.youtube_account_id || '',
              email: ch.email || ''
            })),
            ...unattachedChannels
          ];
          
          setAllChannels(combinedChannels);
          
          // Set initially selected channels (all attached channels)
          const selectedIds = attachedChannels.map(ch => ch.id);
          setSelectedChannelIds(selectedIds);
        } catch (err) {
          console.error("Error fetching channels:", err);
          toast.error("Failed to fetch channels");
        } finally {
          setIsLoading(false);
        }
      };

      fetchChannels();
    }, []);

    const handleAddSelectedChannels = async () => {
      try {
        // Get all selected channels with their full data
        const selectedChannelData = allChannels
          .filter(ch => selectedChannelIds.includes(ch.id))
          .map(ch => ({
            id: ch.id,
            name: ch.name,
            iconUrl: ch.iconUrl,
            email: ch.email || '',
            youtube_account_id: ch.youtube_account_id || ''
          }));

        console.log('Sending channel data:', selectedChannelData);

        await api.post("/api/youtube/add-channels", {
          channels: selectedChannelData
        });
        
        // Refresh the sidebar data
        const res = await api.get("/sidebar-data");
        if (res.data?.channels?.length > 0) {
          onChannelClick?.(res.data.channels[0]); // Select the first channel
        }
        
        setAddChannelDialogOpen(false);
        toast.success("Channels updated successfully");
      } catch (err) {
        console.error("Error updating channels:", err);
        toast.error("Failed to update channels");
      }
    };

    return (
      <Dialog open={addChannelDialogOpen} onClose={() => setAddChannelDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Manage YouTube Channels</DialogTitle>
        <DialogContent>
          <Typography variant="subtitle2" color="text.secondary" mb={2}>
            Select or deselect channels to manage your dashboard
          </Typography>
          <div className="grid gap-4 py-4">
            {isLoading ? (
              <Box display="flex" justifyContent="center" alignItems="center" minHeight={120}>
                <CircularProgress />
              </Box>
            ) : allChannels.length === 0 ? (
              <Box textAlign="center" py={4}>
                <Typography color="text.secondary">No channels found</Typography>
              </Box>
            ) : (
              <List>
                {allChannels.map((channel) => (
                  <ListItem
                    key={channel.id}
                    button="true"
                    onClick={() => {
                      setSelectedChannelIds((prev) =>
                        prev.includes(channel.id)
                          ? prev.filter((id) => id !== channel.id)
                          : [...prev, channel.id]
                      );
                    }}
                    selected={selectedChannelIds.includes(channel.id)}
                  >
                    <Checkbox
                      checked={selectedChannelIds.includes(channel.id)}
                      tabIndex={-1}
                      disableRipple
                    />
                    <ListItemAvatar>
                      <Avatar src={channel.iconUrl}>{channel.name?.[0]}</Avatar>
                    </ListItemAvatar>
                    <ListItemText
                      primary={channel.name}
                      secondary={channel.email}
                    />
                  </ListItem>
                ))}
              </List>
            )}
            <Box textAlign="center" py={4}>
                <Button
                  variant="outlined"
                  color="primary"
                  sx={{ mt: 2 }}
                  onClick={handleAddYouTubeAccount}
                >
                  Add YouTube account
                </Button>
              </Box>
          </div>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setAddChannelDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleAddSelectedChannels} variant="contained">
            Save Changes
          </Button>
        </DialogActions>
      </Dialog>
    );
  };

  return (
    <Box sx={{ height: "100%", display: "flex", flexDirection: "column", bgcolor: "#f5f6fa", p: 2 }}>
      {/* Channels Section */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
        <Typography fontWeight={700} fontSize={15}>
          CHANNELS
        </Typography>
        <Box>
          {role === "owner" && (
            <>
              <IconButton size="small" onClick={handleAddChannel} sx={{ mr: 1 }}>
                <AddIcon />
              </IconButton>
              <IconButton size="small" onClick={() => setUploadDialogOpen(true)}>
                <VideoUploadIcon />
              </IconButton>
            </>
          )}
        </Box>
      </Box>
      <Paper
        sx={{
          display: "flex",
          alignItems: "center",
          mb: 1,
          px: 1,
          py: 0.5,
          borderRadius: 2,
          boxShadow: 0,
          bgcolor: "#fff",
        }}
      >
        <SearchIcon fontSize="small" sx={{ mr: 1 }} />
        <InputBase
          placeholder="search"
          value={search.channels}
          onChange={e => setSearch(s => ({ ...s, channels: e.target.value }))}
          sx={{ flex: 1, fontSize: 14 }}
        />
      </Paper>
      <List dense sx={{ maxHeight: 120, overflow: "auto", mb: 2 }}>
        {(channels || [])
          .filter(item => item.name?.toLowerCase().includes(search.channels.toLowerCase()))
          .map(item => (
            <ListItem
              key={item.id}
              button="true"
              onClick={() => onChannelClick && onChannelClick(item)}
              sx={{ borderRadius: 2, mb: 0.5 }}
            >
              <ListItemAvatar>
                <Avatar src={item.iconUrl}>{item.name?.[0]?.toUpperCase() || "?"}</Avatar>
              </ListItemAvatar>
              <ListItemText primary={item.name} />
            </ListItem>
          ))}
      </List>
      {/* Editors Section */}
      <Divider sx={{ my: 1 }} />
      <Typography fontWeight={700} fontSize={15} mb={1}>
        {role === "owner" ? "EDITORS" : "CLENTS"}
      </Typography>
      <Paper
        sx={{
          display: "flex",
          alignItems: "center",
          mb: 1,
          px: 1,
          py: 0.5,
          borderRadius: 2,
          boxShadow: 0,
          bgcolor: "#fff",
        }}
      >
        <SearchIcon fontSize="small" sx={{ mr: 1 }} />
        <InputBase
          placeholder="search"
          value={search.clients}
          onChange={e => setSearch(s => ({ ...s, clients: e.target.value }))}
          sx={{ flex: 1, fontSize: 14 }}
        />
        {/* Show Add Editor for both owner and editor */}
        <IconButton size="small" onClick={onAddClient}>
          <AddIcon />
        </IconButton>
      </Paper>
      <List dense sx={{ maxHeight: 120, overflow: "auto" }}>
        {(clients || [])
          .filter(item => item.name?.toLowerCase().includes(search.clients.toLowerCase()))
          .map(item => (
            <ListItem
              key={item.id}
              button="true"
              onClick={() => onClientClick && onClientClick(item)}
              sx={{ borderRadius: 2, mb: 0.5 }}
            >
              <ListItemAvatar>
                <Avatar>{item.name?.[0]?.toUpperCase() || "?"}</Avatar>
              </ListItemAvatar>
              <ListItemText primary={item.name} />
            </ListItem>
          ))}
      </List>

      {/* Add Channel Dialog */}
      <AddChannelDialog />
      
      {/* Video Upload Dialog */}
      <VideoUpload 
        open={uploadDialogOpen}
        onClose={() => setUploadDialogOpen(false)}
        onSuccess={() => {
          toast.success("Video uploaded successfully");
          // You might want to refresh the project list here
        }}
      />
    </Box>
  );
} 