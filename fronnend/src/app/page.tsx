'use client'
import React, { useState } from 'react';
import { motion, AnimatePresence } from "framer-motion";
import { VideoIcon, UserPlus, ChevronRight, Star, Command, X } from "lucide-react";
import Link from "next/link";

export default function Home() {
  const [isFeatureModalOpen, setIsFeatureModalOpen] = useState(false);
  const [activeFeature, setActiveFeature] = useState(null);

  const features = [
    {
      icon: <VideoIcon className="w-12 h-12 text-retroOrange" />,
      title: "HD Video Calls",
      description: "Crystal clear video with retro-inspired filters and effects."
    },
    {
      icon: <Command className="w-12 h-12 text-retroPink" />,
      title: "Custom Rooms",
      description: "Create unique virtual spaces with nostalgic themes."
    },
    {
      icon: <Star className="w-12 h-12 text-retroPurple" />,
      title: "Fun Interactions",
      description: "Engage with interactive emotes and background effects."
    }
  ];

  return (
    <div className="min-h-screen bg-gradient-to-br from-amber-200 via-pink-500 to-purple-500 overflow-hidden relative">
      {/* Header */}
      <header className="absolute top-0 left-0 w-full p-6 flex justify-between items-center z-20">
        <Link href="/">
          <motion.div 
            className="text-3xl font-bold text-orange-600 flex items-center gap-2"
            whileHover={{ scale: 1.05 }}
          >
            <Command className="w-8 h-8" />
            RetroConfer
          </motion.div>
        </Link>
        <div className="flex items-center gap-4">
          <motion.button 
            className="px-4 py-2 bg-orange-600 text-white font-bold rounded-lg shadow-md hover:bg-orange-700 transition-colors"
            whileHover={{ scale: 1.05 }}
          >
            Sign In
          </motion.button>
        </div>
      </header>

      {/* Main Content */}
      <div className="h-screen  mt-20 flex flex-col items-center justify-center text-center px-4">
        <motion.div
          initial={{ opacity: 0, y: -50 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8 }}
          className="mb-12"
        >
          <h1 className="text-5xl font-bold text-white mb-4">
            Connect. Collaborate. Celebrate.
          </h1>
          <p className="text-xl text-white max-w-2xl mx-auto">
            Video Conferencing powered by Golang!
          </p>
        </motion.div>

        {/* Action Buttons */}
        <div className="flex gap-6">
          <Link href="/create-room">
            <motion.button
              className="flex items-center gap-2 px-6 py-3 bg-orange-600 text-white font-bold rounded-lg shadow-lg hover:bg-orange-700 transition-colors"
              whileHover={{ scale: 1.1 }}
            >
              <VideoIcon />
              Create Room
            </motion.button>
          </Link>
          <Link href="/join-room">
            <motion.button
              className="flex items-center gap-2 px-6 py-3 bg-pink-500 text-white font-bold rounded-lg shadow-lg hover:bg-pink-600 transition-colors"
              whileHover={{ scale: 1.1 }}
            >
              <UserPlus />
              Join Room
            </motion.button>
          </Link>
        </div>

        {/* Features Section */}
        <div className="mt-16 flex justify-center gap-8">
          {features.map((feature, index) => (
            <motion.div
              key={index}
              className="bg-white/20 backdrop-blur-sm rounded-xl p-6 w-64 text-center"
              whileHover={{ 
                scale: 1.05,
                boxShadow: "0 10px 20px rgba(0,0,0,0.2)"
              }}
              onClick={() => {
                setActiveFeature(feature);
                setIsFeatureModalOpen(true);
              }}
            >
              <div className="flex justify-center mb-4">
                {feature.icon}
              </div>
              <h3 className="text-xl font-bold text-white mb-2">{feature.title}</h3>
              <p className="text-white/80 text-sm">{feature.description}</p>
            </motion.div>
          ))}
        </div>

        {/* Feature Modal */}
        {isFeatureModalOpen && activeFeature && (
          <div 
            className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
            onClick={() => setIsFeatureModalOpen(false)}
          >
            <motion.div
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              className="bg-white rounded-2xl p-8 max-w-md w-full relative"
              onClick={(e) => e.stopPropagation()}
            >
              <button 
                className="absolute top-4 right-4 text-gray-500 hover:text-gray-800"
                onClick={() => setIsFeatureModalOpen(false)}
              >
                <X className="w-6 h-6" />
              </button>
              <div className="flex justify-center mb-6">
                {activeFeature.icon}
              </div>
              <h2 className="text-2xl font-bold text-center mb-4">
                {activeFeature.title}
              </h2>
              <p className="text-center text-gray-600">
                {activeFeature.description}
              </p>
            </motion.div>
          </div>
        )}
      </div>

      {/* Background Animation */}
      <motion.div
        className="absolute inset-0 -z-10 opacity-50"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 1.5 }}
      >
        <svg
          viewBox="0 0 800 600"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className="w-full h-full"
        >
          <circle cx="200" cy="200" r="150" fill="#FB923C" />
          <circle cx="600" cy="400" r="200" fill="#A78BFA" />
          <circle cx="400" cy="100" r="100" fill="#F472B6" />
        </svg>
      </motion.div>
    </div>
  );
}