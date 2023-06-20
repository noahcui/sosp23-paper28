#!/usr/bin/env python3

import argparse
import pandas as pd
import matplotlib.pyplot as plt

def main(input_file, column, output_file):
    # Load the CSV data into a pandas DataFrame
    df = pd.read_csv(input_file)

    # Convert the Unix timestamp to datetime and set it as the index
    df['Time'] = pd.to_datetime(df.iloc[:, 0], unit='s')

    # Subtract the first timestamp to start from 0
    df['Time'] = df['Time'] - df['Time'].iloc[0]
    df.set_index('Time', inplace=True)

    # Compute the average per second
    df[column] = df[column].resample('S').mean()

    # Plot the data
    plt.plot(df.index.total_seconds(), df[column])  # Convert datetime index to seconds
    plt.xlabel('Time (s)')
    plt.ylabel(column)
    plt.title(f'Average {column} over time')

    # Save the figure to a PDF file
    plt.savefig(output_file)

if __name__ == '__main__':
    # Setup the argument parser
    parser = argparse.ArgumentParser(description='Plot data from a CSV file.')
    parser.add_argument('-input', type=str, required=True, help='Input CSV file.')
    parser.add_argument('-y', type=str, required=True, help='Column to plot.')
    parser.add_argument('-outfile', type=str, required=True, help='Output PDF file.')

    # Parse the arguments
    args = parser.parse_args()

    # Run the main function
    main(args.input, args.y, args.outfile)
