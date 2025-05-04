import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  Avatar,
  Checkbox,
  Button,
  CircularProgress,
  Box,
  Typography,
} from '@mui/material';
import { toast } from 'sonner';
import axios from 'axios';

export default function OauthCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [channels, setChannels] = useState([]);
  const [selectedChannels, setSelectedChannels] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [open, setOpen] = useState(true);
  const [youtubeAccountId, setYoutubeAccountId] = useState('');

  useEffect(() => {
    const status = searchParams.get('status');
    const channelsParam = searchParams.get('channels');
    const email = searchParams.get('email');
    const userId = searchParams.get('user_id');

    if (status !== 'success' || !channelsParam || !email || !userId) {
      toast.error('Invalid callback parameters');
      navigate('/');
      return;
    }

    try {
      const decodedChannels = JSON.parse(decodeURIComponent(channelsParam));
      setChannels(decodedChannels.map(ch => ({
        ...ch,
        email: email,
        youtube_account_id: userId
      })));
      setSelectedChannels(decodedChannels.map(ch => ch.id));
    } catch (error) {
      console.error('Error parsing channels:', error);
      toast.error('Failed to parse channels data');
      navigate('/');
    }
  }, [searchParams, navigate]);

  const handleChannelToggle = (channelId) => {
    setSelectedChannels(prev =>
      prev.includes(channelId)
        ? prev.filter(id => id !== channelId)
        : [...prev, channelId]
    );
  };

  const handleSubmit = async () => {
    if (selectedChannels.length === 0) {
      toast.error('Please select at least one channel');
      return;
    }

    setIsLoading(true);
    try {
      const selectedChannelData = channels
        .filter(ch => selectedChannels.includes(ch.id));

      await axios.post('http://localhost:8080/api/youtube/add-channels', {
        channels: selectedChannelData
      }, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
      toast.success('Channels added successfully');
      setOpen(false);
      navigate('/dashboard');
    } catch (error) {
      console.error('Error adding channels:', error);
      toast.error('Failed to add channels');
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    setOpen(false);
    navigate('/dashboard');
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Select YouTube Channels</DialogTitle>
      <DialogContent>
        <Typography mb={2}>Select channels to add to your dashboard:</Typography>
        {channels.length === 0 ? (
          <Typography color="text.secondary">No channels found.</Typography>
        ) : (
          <List>
            {channels.map((channel) => (
              <ListItem
                key={channel.id}
                button="true"
                onClick={() => handleChannelToggle(channel.id)}
                selected={selectedChannels.includes(channel.id)}
                sx={{ borderRadius: 2, mb: 0.5 }}
              >
                <Checkbox
                  checked={selectedChannels.includes(channel.id)}
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
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>Cancel</Button>
        <Button
          onClick={handleSubmit}
          variant="contained"
          disabled={isLoading || selectedChannels.length === 0}
        >
          {isLoading ? <CircularProgress size={20} /> : 'Add Selected'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
